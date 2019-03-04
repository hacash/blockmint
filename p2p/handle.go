package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
	"sync"
)

const (

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

type txsync struct {
	p   *peer
	txs []block.Transaction
}

type minersync struct {
	p *peer
}

type ProtocolManager struct {
	TxsCh               chan []block.Transaction          // 交易广播
	DiscoveryNewBlockCh chan miner.DiscoveryNewBlockEvent // 挖出新区块广播

	maxPeers     int // 最大节点数量
	peers        *peerSet
	SubProtocols []p2p.Protocol

	txpool service.TxPool
	miner  *miner.HacashMiner

	// 同步交易
	txsyncCh    chan *txsync
	minersyncCh chan *minersync
	onsyncminer bool // 是否正在同步状态

	// 全网最新区块状态
	//netBestHeadHash []byte
	//netBestHeadHeight uint64

}

var (
	globalInstanceProtocolManagerMutex sync.Mutex
	globalInstanceProtocolManager      *ProtocolManager = nil
)

func GetGlobalInstanceProtocolManager() *ProtocolManager {
	globalInstanceProtocolManagerMutex.Lock()
	defer globalInstanceProtocolManagerMutex.Unlock()
	if globalInstanceProtocolManager == nil {
		globalInstanceProtocolManager = NewProtocolManager()
	}
	return globalInstanceProtocolManager
}

func NewProtocolManager() *ProtocolManager {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		maxPeers:    25,
		peers:       newPeerSet(),
		txpool:      txpool.GetGlobalInstanceMemTxPool(),
		miner:       miner.GetGlobalInstanceHacashMiner(),
		onsyncminer: false,
	}

	manager.SubProtocols = make([]p2p.Protocol, 0, 1)
	manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
		Name:    "hacash",
		Version: 1,
		Length:  200,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			//fmt.Println("Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error")
			peer := manager.newPeer(p, rw)
			return manager.handle(peer)
		},
	})

	return manager
}

func (pm *ProtocolManager) Start(maxPeers int) {
	if maxPeers > 0 {
		pm.maxPeers = maxPeers
	}

	// 新区块广播
	pm.DiscoveryNewBlockCh = make(chan miner.DiscoveryNewBlockEvent, 256)
	pm.miner.SubscribeDiscoveryNewBlock(pm.DiscoveryNewBlockCh)
	go pm.blockBroadcastLoop()

	// broadcast transactions
	pm.TxsCh = make(chan []block.Transaction, txChanSize)
	pm.txpool.SubscribeNewTx(pm.TxsCh)
	go pm.txBroadcastLoop()

	// sync transactions in pool
	pm.txsyncCh = make(chan *txsync)
	go pm.txsyncLoop()

	// 矿工状态同步
	pm.minersyncCh = make(chan *minersync)
	go pm.minersyncLoop()

}

func (pm *ProtocolManager) newPeer(p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(p, rw)
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	//log.Debug("Removing Hacash peer", "peer", id)
	if err := pm.peers.Unregister(id); err != nil {
		fmt.Errorf("Peer removal failed peer %s err %s", id, err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (pm *ProtocolManager) blockBroadcastLoop() {
	for {
		select {
		case blk := <-pm.DiscoveryNewBlockCh:
			data := MsgDataNewBlock{
				block:  blk.Block,
				Height: blk.Block.GetHeight(),
				Datas:  string(blk.Bodys),
			}
			// 广播区块
			pm.BroadcastBlock(&data)
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case txs := <-pm.TxsCh:
			pm.BroadcastTxs(txs)
		}
	}
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(newBlock *MsgDataNewBlock) {
	hash := newBlock.block.Hash()
	peers := pm.peers.PeersWithoutBlock(hash)

	// Otherwise if the block is indeed in out own chain, announce it
	for _, peer := range peers {
		peer.AsyncSendNewBlockHash(newBlock)
	}
	if len(peers) > 0 {
		fmt.Println("broadcast block", "height", newBlock.Height, "hash", hex.EncodeToString(hash), "recipients", len(peers))
	}

}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs []block.Transaction) {

	var txset = make(map[*peer][]block.Transaction)

	// Broadcast transactions to a batch of peers not knowing about it
	for _, tx := range txs {
		peers := pm.peers.PeersWithoutTx(tx.Hash())
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		fmt.Println("Broadcast transaction", "hash", hex.EncodeToString(tx.Hash()), "recipients", len(peers))
	}
	//
	for peer, txs := range txset {
		peer.AsyncSendTransactions(txs)
	}
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncTransactions(p *peer) {
	var txs = make([]block.Transaction, 0, 16)
	// read txs
	if len(txs) == 0 {
		return
	}
	pm.txsyncCh <- &txsync{p, txs}
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncMinerStatus(p *peer) {
	pm.minersyncCh <- &minersync{p}
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) DoSyncMinerStatus(p *peer) {

	//fmt.Println("func DoSyncMinerStatus")

	if pm.onsyncminer {
		return
	}
	if pm.peers.Len() < 1 {
		return // 最少连接个节点才能同步状态
	}
	// 判断是否完成同步
	best := pm.peers.BestPeer()
	if best == nil {
		return // not Completed
	}
	_, tarhei := best.Head()
	minerdb := pm.miner
	fromhei := minerdb.State.CurrentHeight() + 1
	if tarhei > fromhei {
		pm.onsyncminer = true // 开始同步
		fmt.Printf("sync blocks from height %d ...\n", fromhei)
		p2p.Send(best.rw, GetSyncBlocksMsg, MsgDataGetSyncBlocks{
			fromhei,
		})
	} else {
		pm.onsyncminer = false
		minerdb.CanStart() // 可以开始挖矿
	}
}

func (pm *ProtocolManager) txsyncLoop() {

	select {
	case s := <-pm.txsyncCh:
		s.p.SendTransactions(s.txs)
	}
}

func (pm *ProtocolManager) minersyncLoop() {

	select {
	case m := <-pm.minersyncCh:
		pm.DoSyncMinerStatus(m.p) // 同步矿工状态
	}
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	//fmt.Println("Hacash peer to do handshake...", "name", p.Name())

	// Execute the Hacash handshake
	if err := p.Handshake(); err != nil {
		fmt.Println("Hacash handshake failed", "err", err)
		return err
	}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		fmt.Println("Hacash peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)

	// 同步矿工状态
	pm.syncMinerStatus(p)

	fmt.Println("Hacash peer connected", "name:", p.Name())

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			fmt.Println("Hacash message handling failed", "err", err)
			//return err
			return nil
		}
	}

}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed

	//fmt.Println("handleMsg +++++++++++")

	msg, err := p.rw.ReadMsg()

	//fmt.Println("(pm *ProtocolManager) handleMsg(p *peer) error handleMsg handleMsg handleMsg handleMsg handleMsg ", msg.Code)
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return fmt.Errorf("%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return fmt.Errorf("uncontrolled status message")

	case msg.Code == TxMsg:

		//fmt.Println("case msg.Code == TxMsg: ")

		// Transactions can be processed, parse all of them and deliver to the pool
		//var txs []block.Transaction
		var sdtxs []string

		if err := msg.Decode(&sdtxs); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}

		//fmt.Println(" peer from ", p.Info())
		//fmt.Println("myMessage ::::::::: ", sdtxs)

		for i, txstr := range sdtxs {
			tx, _, e0 := blocks.ParseTransaction([]byte(txstr), 0)
			// Validate and mark the remote transaction
			if e0 != nil {
				return fmt.Errorf("transaction format error", i)
			}
			p.MarkTransaction(tx.Hash())
			// 加入我的交易池
			pe := pm.txpool.AddTx(tx)
			if pe != nil {
				fmt.Println("pm.txpool.AddTx error:", pe)
			}
			//fmt.Println("pm.txpool.AddTx(tx) 加入我的交易池  txtxtxtxtxtxtxtxtxtx ", hex.EncodeToString(tx.HashNoFee()))
		}

	case msg.Code == GetSyncBlocksMsg:
		// 请求同步区块消息
		fmt.Println("case msg.Code == GetSyncBlocksMsg: ")
		var data MsgDataGetSyncBlocks
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		fmt.Println("msg.Decode(&data)")
		minerdb := pm.miner
		if minerdb.State.CurrentHeight() <= data.StartHeight {
			return nil // 不能提供
		}
		fmt.Println("miner.GetGlobalInstanceHacashMiner()")
		// 查询区块并返回
		go func() {
			bkdb := store.GetGlobalInstanceBlocksDataStore()
			var blocks bytes.Buffer
			var blocklen = 0
			var blocksize = 0
			//fmt.Println("for height:=data.StartHeight; height <= minerdb.State.CurrentHeight(); height++ { ", data.StartHeight, minerdb.State.CurrentHeight())
			for height := data.StartHeight; height <= minerdb.State.CurrentHeight(); height++ {
				blkbytes, e := bkdb.GetBlockBytesByHeight(height, true, true)
				size := len(blkbytes)
				if e != nil || size == 0 {
					fmt.Printf(" bkdb.GetBlockBytesByHeight: %d, len: %d \n", height, size)
					return // 不能提供
				}
				fmt.Printf("blocks = append(blocks, string(blkbytes)), height=%d, length=%d, string=%s \n", height, len(blkbytes), hex.EncodeToString(blkbytes))
				blocks.Write(blkbytes)
				blocksize += size
				blocklen++
				if blocklen >= 100 || blocksize >= 1024*1024 {
					fmt.Println("totalsize >= 512KB break ")
					break
				}
			}
			if blocksize == 0 {
				fmt.Println("blocks.Len() == 0")
				return
			}
			// 发送
			sdmsgdt := MsgDataSyncBlocks{
				FromHeight: data.StartHeight,
				ToHeight:   data.StartHeight + uint64(blocklen) - 1,
				Datas:      blocks.String(),
			}
			fmt.Printf("Send SyncBlocks, FromHeight: %d, ToHeight: %d \n", sdmsgdt.FromHeight, sdmsgdt.ToHeight)
			go p2p.Send(p.rw, SyncBlocksMsg, sdmsgdt)
		}()

	case msg.Code == SyncBlocksMsg:
		// 请求同步区块消息
		//fmt.Println("case msg.Code == SyncBlocksMsg: ")
		var data MsgDataSyncBlocks
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		// 检查
		minerdb := pm.miner
		if minerdb.State.CurrentHeight()+1 != data.FromHeight {
			return nil
		}

		//fmt.Printf("SyncBlocksMsg, FromHeight: %d, ToHeight: %d,  \n", data.FromHeight, data.ToHeight)
		go func() {
			// 解包区块，依次插入
			stuffbytes := []byte(data.Datas)
			//fmt.Println(hex.EncodeToString(stuffbytes))
			seek := uint32(0)
			var insertok error
			var blk block.Block
			for {
				//fmt.Println("seek", seek)
				blk, seek, insertok = minerdb.ArrivedNewBlockToUpdate(stuffbytes, seek)
				if insertok != nil {
					fmt.Printf("peer %s give error block data", p.Name(), insertok)
					return
				}
				// 标记已知
				p.MarkBlock(blk.Hash())
				if seek >= uint32(len(stuffbytes)) {
					break // finish
				}
			}
			fmt.Printf("sync blocks from height: %d, to height: %d, ok\n", data.FromHeight, data.ToHeight)
			// 判断是否完成同步
			_, tarhei := pm.peers.BestPeer().Head()
			if tarhei <= data.ToHeight {
				pm.onsyncminer = false // 结束同步
				minerdb.CanStart()     // 可以开始挖矿
				return                 // 同步完成
			}
			// 再次发起同步请求
			fromhei := data.ToHeight + 1
			fmt.Printf("sync blocks from height %d ...\n", fromhei)
			p2p.Send(p.rw, GetSyncBlocksMsg, MsgDataGetSyncBlocks{
				fromhei,
			})
		}()

	case msg.Code == NewBlockExcavateMsg:
		// 新区块被挖出
		fmt.Println("case msg.Code == NewBlockExcavateMsg: ")
		if pm.onsyncminer {
			return nil // 正在同步 忽略新挖出区块
		}
		// 新区块被挖出
		fmt.Println("var data MsgDataNewBlock")
		var data MsgDataNewBlock
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		fmt.Println("err := msg.Decode(&data);", data.Height)
		// 检查
		minerdb := pm.miner
		mytarhei := minerdb.State.CurrentHeight() + 1
		if mytarhei > data.Height {
			return nil
		}
		if mytarhei < data.Height {
			go func() {
				// 发起同步
				pm.onsyncminer = true // 开始同步
				fmt.Printf("sync blocks from height %d ...\n", mytarhei)
				p2p.Send(p.rw, GetSyncBlocksMsg, MsgDataGetSyncBlocks{
					mytarhei,
				})
			}()
			return nil
		}
		fmt.Printf("blk, _, inserterr := minerdb.ArrivedNewBlockToUpdate, height: %d \n", data.Height)
		go func() {
			// 插入
			blk, _, inserterr := minerdb.ArrivedNewBlockToUpdate([]byte(data.Datas), 0)
			if inserterr != nil {

				return
			}
			p.MarkBlock(blk.Hash())
			fmt.Println("重新开始挖矿 minerdb.CanStart()")
			// 重新开始挖矿
			minerdb.CanStart()
			// 广播区块
			data.block = blk
			go pm.BroadcastBlock(&data)
			fmt.Println("go pm.BroadcastBlock(&data)")
		}()

	default:
		return fmt.Errorf("%v", msg.Code)
	}

	return nil
}
