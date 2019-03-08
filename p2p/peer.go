package p2p

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/hacash/blockmint/types/block"
	"sync"
	"time"
)

const (
	maxKnownTxs        = 32768 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownBlocks     = 1024  // Maximum block hashes to keep in the known list (prevent DOS)
	maxQueuedTxs       = 128
	maxQueuedNewBlocks = 64
	maxQueuedHeight    = 64

	handshakeTimeout = 5 * time.Second
)

///////

type peer struct {
	id string

	*p2p.Peer
	rw p2p.MsgReadWriter

	knownTxs    mapset.Set // Set of transaction hashes known to be known by this peer
	knownBlocks mapset.Set // Set of block hashes known to be known by this peer

	queuedTxs      chan []block.Transaction // Queue of transactions to broadcast to the peer
	queuedNewBlock chan *MsgDataNewBlock    // Queue of blocks to announce to the peer
	queuedHeight   chan uint64              // Queue of height

	headBlockHash []byte // 当前所在区块hash
	knownHeight   uint64

	lock sync.RWMutex

	term chan struct{} // Termination channel to stop the broadcaster

}

func newPeer(p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return &peer{
		Peer:           p,
		rw:             rw,
		id:             fmt.Sprintf("%x", p.ID().Bytes()[:8]),
		knownTxs:       mapset.NewSet(),
		knownBlocks:    mapset.NewSet(),
		queuedTxs:      make(chan []block.Transaction, maxQueuedTxs),
		queuedNewBlock: make(chan *MsgDataNewBlock, maxQueuedNewBlocks),
		queuedHeight:   make(chan uint64, maxQueuedHeight),

		term: make(chan struct{}),
	}
}

// broadcast is a write loop that multiplexes block propagations, announcements
// and transaction broadcasts into the remote peer. The goal is to have an async
// writer that does not lock up node internals.
func (p *peer) broadcast() {
	for {
		select {
		case txs := <-p.queuedTxs:
			if err := p.SendTransactions(txs); err != nil {
				return
			}
			// fmt.Println("Broadcast transactions", "count", len(txs))

		case data := <-p.queuedNewBlock:
			if err := p.SendNewBlock(data); err != nil {
				return
			}

		case height := <-p.queuedHeight:
			if err := p.SendHigherHeight(height); err != nil {
				return
			}
			// p.Log().Trace("send new block", "height", data.block.GetHeight(), "hash", hex.EncodeToString(data.block.Hash()))

		case <-p.term:
			return
		}
	}
}

// close signals the broadcast goroutine to terminate.
func (p *peer) close() {
	close(p.term)
}

// Head retrieves a copy of the current headBlockHash hash and total difficulty of the
// peer.
func (p *peer) Head() (hash []byte, height uint64) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	height = p.knownHeight
	copy(hash[:], p.headBlockHash[:])
	return hash, height
}

// MarkBlock marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
func (p *peer) MarkBlock(block block.Block) {
	// If we reached the memory allowance, drop a previously known block hash
	for p.knownBlocks.Cardinality() >= maxKnownBlocks {
		p.knownBlocks.Pop()
	}
	p.knownBlocks.Add(string(block.Hash()))
	if p.knownHeight < block.GetHeight() {
		p.knownHeight = block.GetHeight()
		p.headBlockHash = block.Hash()
	}
}

// MarkTransaction marks a transaction as known for the peer, ensuring that it
// will never be propagated to this particular peer.
func (p *peer) MarkTransaction(hash []byte) {
	// If we reached the memory allowance, drop a previously known transaction hash
	for p.knownTxs.Cardinality() >= maxKnownTxs {
		p.knownTxs.Pop()
	}
	p.knownTxs.Add(string(hash))
}

// SendNewBlockHashes announces the availability of a number of blocks through
// a hash notification.
func (p *peer) SendNewBlock(newBlock *MsgDataNewBlock) error {
	return p2p.Send(p.rw, NewBlockExcavateMsg, newBlock)
}

// SendTransactions sends transactions to the peer and includes the hashes
// in its transaction hash set for future reference.
func (p *peer) SendTransactions(txs []block.Transaction) error {
	var sdtxs []string
	for _, tx := range txs {
		p.knownTxs.Add(string(tx.Hash()))
		txbys, _ := tx.Serialize()
		sdtxs = append(sdtxs, string(txbys))
	}
	return p2p.Send(p.rw, TxMsg, sdtxs)
}

func (p *peer) SendHigherHeight(height uint64) error {
	msgdata := MsgDataHeightHigher{Height: height}
	return p2p.Send(p.rw, HeightHigherMsg, msgdata)
}

func (p *peer) AsyncSendNewBlock(data *MsgDataNewBlock) {
	select {
	case p.queuedNewBlock <- data:
		p.knownBlocks.Add(string(data.block.Hash()))
	default:
		p.Log().Debug("Dropping block announcement", "height", data.block.GetHeight(), "hash", data.block.Hash())
	}
}

// AsyncSendTransactions queues list of transactions propagation to a remote
// peer. If the peer's broadcast queue is full, the event is silently dropped.
func (p *peer) AsyncSendTransactions(txs []block.Transaction) {
	select {
	case p.queuedTxs <- txs:
		for _, tx := range txs {
			p.knownTxs.Add(string(tx.Hash()))
		}
	default:
		p.Log().Debug("Dropping transaction propagation", "count", len(txs))
	}
}

func (p *peer) AsyncSendHigherHeight(height uint64) {
	select {
	case p.queuedHeight <- height:
		if height > p.knownHeight {
			p.knownHeight = height
		}
	default:
		p.Log().Debug("Dropping height propagation", "height", height)
	}

}

// Handshake executes the eth protocol handshake, negotiating version number,
// network IDs, difficulties, headBlockHash and genesis blocks.
func (p *peer) Handshake() error {
	errc := make(chan error, 2)
	// Send out own handshake in a new thread
	selfstatus := CreateHandShakeStatusData()
	var status handShakeStatusData
	go func() {
		errc <- p2p.Send(p.rw, StatusMsg, selfstatus)
	}()
	go func() {
		errc <- p.readStatus(&selfstatus, &status)
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				//fmt.Println("case err := <-errc:", err)
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	//fmt.Println("(p *peer) Handshake() error   nononononon error")
	return nil
}

func (p *peer) readStatus(selfstatus, status *handShakeStatusData) error {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return fmt.Errorf("first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return fmt.Errorf("%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(status); err != nil {
		return fmt.Errorf("msg %v: %v", msg, err)
	}

	//fmt.Println(" readStatus =========== handShakeStatusData from ", p.Name())
	stok := selfstatus.Confirm(status)
	if stok != nil {
		return stok
	}

	p.headBlockHash = status.CurrentBlockHash
	p.knownHeight = status.CurrentBlockHeight

	//fmt.Println(" selfstatus.Confirm(status) Completed", p.Name())

	return nil
}
