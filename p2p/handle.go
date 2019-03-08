package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
	"sync"
	"time"
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

type heightsync struct {
	p      *peer
	height uint64
}

type ProtocolManager struct {
	TxsCh               chan []block.Transaction          // 交易广播
	DiscoveryNewBlockCh chan miner.DiscoveryNewBlockEvent // 挖出新区块广播
	HeightHighCh        chan uint64

	maxPeers     int // 最大节点数量
	peers        *peerSet
	SubProtocols []p2p.Protocol

	txpool service.TxPool
	miner  *miner.HacashMiner

	// 同步交易
	txsyncCh       chan *txsync
	minersyncCh    chan *minersync
	heightsyncCh   chan *heightsync
	onsyncminer    bool // 是否正在同步状态
	onheighthigher bool // 是否是发现了新高度状态

	Log log.Logger
}

var (
	globalInstanceProtocolManagerMutex sync.Mutex
	globalInstanceProtocolManager      *ProtocolManager = nil
)

func GetGlobalInstanceProtocolManager() *ProtocolManager {
	globalInstanceProtocolManagerMutex.Lock()
	defer globalInstanceProtocolManagerMutex.Unlock()
	if globalInstanceProtocolManager == nil {
		lg := config.GetGlobalInstanceLogger()
		globalInstanceProtocolManager = NewProtocolManager(lg)
	}
	return globalInstanceProtocolManager
}

func NewProtocolManager(log log.Logger) *ProtocolManager {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		maxPeers:       25,
		peers:          newPeerSet(),
		txpool:         txpool.GetGlobalInstanceMemTxPool(),
		miner:          miner.GetGlobalInstanceHacashMiner(),
		onsyncminer:    false,
		onheighthigher: false,
		Log:            log,
	}

	manager.SubProtocols = make([]p2p.Protocol, 0, 1)
	manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
		Name:    "hacash",
		Version: 1,
		Length:  200,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			//fmt.Println("manager.SubProtocols Run: peer ", p.Name())
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

	pm.Log.News("protocal manager start")

	// 新区块广播
	pm.DiscoveryNewBlockCh = make(chan miner.DiscoveryNewBlockEvent, 100)
	pm.miner.SubscribeDiscoveryNewBlock(pm.DiscoveryNewBlockCh)
	go pm.blockBroadcastLoop()

	// broadcast transactions
	pm.TxsCh = make(chan []block.Transaction, txChanSize)
	pm.txpool.SubscribeNewTx(pm.TxsCh)
	go pm.txBroadcastLoop()

	pm.HeightHighCh = make(chan uint64, 100)
	go pm.heightBroadcastLoop()

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
	pm.Log.Info("removing hacash peer", peer.Name())

	if err := pm.peers.Unregister(id); err != nil {
		pm.Log.Warning("peer removal failed peer id %s err %s", id, err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
	// 如果连接数为零，且不为强制挖矿，则停止挖矿，避免在自己的分支上挖矿
	if pm.peers.Len() == 0 && config.Config.Miner.Forcestart != "true" {
		// 停止挖矿
		pm.Log.Warning("only one peer removed", peer.Name(), "no connect now, stop mining !")
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

func (pm *ProtocolManager) heightBroadcastLoop() {
	for {
		select {
		case height := <-pm.HeightHighCh:
			pm.BroadcastHeight(height)
		}
	}
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(newBlock *MsgDataNewBlock) {
	hash := newBlock.block.Hash()
	peers := pm.peers.PeersWithoutBlock(hash)

	if len(peers) > 0 {
		pm.Log.Info("broadcast block to", len(peers), "peers", "height", newBlock.block.GetHeight(), "hash", hex.EncodeToString(hash))
	}

	// Otherwise if the block is indeed in out own chain, announce it
	for _, peer := range peers {
		peer.AsyncSendNewBlock(newBlock)
	}
	//if len(peers) > 0 {
	//fmt.Println("broadcast block", "height", newBlock.Height, "hash", hex.EncodeToString(hash), "recipients", len(peers))
	//}

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
		if len(peers) > 0 {
			pm.Log.Info("broadcast tx to", len(peers), "peers", "hash", hex.EncodeToString(tx.Hash()))
		}
	}
	//
	for peer, txs := range txset {
		peer.AsyncSendTransactions(txs)
	}
}

func (pm *ProtocolManager) BroadcastHeight(height uint64) {
	peers := pm.peers.PeersWithoutHeight(height)
	if len(peers) > 0 {
		pm.Log.Info("broadcast height to", len(peers), "peers", "height", height)
	}
	for _, peer := range peers {
		peer.AsyncSendHigherHeight(height)
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
	pm.Log.Noise("p2p to sync miner status")
	//fmt.Println("func DoSyncMinerStatus, onsyncminer:", pm.onsyncminer)
	//if pm.onsyncminer {
	//	return // 正在同步
	//}
	_, tarhei := p.Head()
	minerdb := pm.miner
	fromhei := minerdb.State.CurrentHeight() + 1
	if tarhei >= fromhei {
		//fmt.Println("tarhei >= fromhei")
		//fmt.Println("pm.onsyncminer = true")
		pm.onsyncminer = true // 开始同步
		pm.Log.Mark("request sync blocks from height", fromhei)
		go func() {
			p2p.Send(p.rw, GetSyncBlocksMsg, MsgDataGetSyncBlocks{
				fromhei,
			})
		}()
	} else {
		pm.Log.Mark("all block status sync finish, start mining ...")
		pm.onsyncminer = false
		minerdb.StartMining() // 可以开始挖矿
		if pm.onheighthigher == true {
			// 如果是发现了新高度， 则广播通知大家
			pm.onheighthigher = false
			pm.BroadcastHeight(tarhei)
		}
	}
}

func (pm *ProtocolManager) txsyncLoop() {

	for {
		select {
		case s := <-pm.txsyncCh:
			s.p.SendTransactions(s.txs)
		}
	}
}

func (pm *ProtocolManager) minersyncLoop() {

	for {
		select {
		case m := <-pm.minersyncCh:
			pm.DoSyncMinerStatus(m.p) // 同步矿工状态
		}
	}
}

// handle is the callback invoked to manage the life cycle of an eth peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p *peer) error {

	pm.Log.Noise("p2p handle msg from peer", p.Name())

	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted {
		//fmt.Println("pm.peers.Len() >= pm.maxPeers && !p.Peer.Info().Network.Trusted", "p", p.Name())
		return p2p.DiscTooManyPeers
	}
	//fmt.Println("Hacash peer to do handshake...", "name", p.Name())

	// Execute the Hacash handshake
	if err := p.Handshake(); err != nil {
		pm.Log.Attention("p2p handshake failed", "err", err)
		return err
	}
	// Register the peer locally
	if err := pm.peers.Register(p); err != nil {
		pm.Log.Warning("p2p peer registration failed", "err", err)
		return err
	}
	defer pm.removePeer(p.id)

	pm.Log.News("p2p peer connected", "name:", p.Name())

	//fmt.Println("syncTransactions", "name", p.Name())

	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)
	//fmt.Println("syncMinerStatus", "name", p.Name())

	// 同步矿工状态
	pm.syncMinerStatus(p)
	//fmt.Println("Hacash peer connected", "name:", p.Name())

	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {
			//fmt.Println("Hacash message handling failed", "err", err)
			return err
		}
	}

}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed

	//fmt.Printf("handleMsg ++++++++++++++++++++++++++++++++++++++++++++++++++ peer: %s  \n", p.Name())

	msg, err := p.rw.ReadMsg()

	pm.Log.Noise("p2p peer", p.Name(), "push msg code", msg.Code)

	//fmt.Printf("p.rw.ReadMsg ------ msg.Code ==  %d \n", msg.Code)

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
				// fmt.Println("pm.txpool.AddTx error:", pe)
			}
			//fmt.Println("pm.txpool.AddTx(tx) 加入我的交易池  txtxtxtxtxtxtxtxtxtx ", hex.EncodeToString(tx.HashNoFee()))
		}

	case msg.Code == GetSyncBlocksMsg:
		// 请求同步区块消息
		//fmt.Println("case msg.Code == GetSyncBlocksMsg: ")
		var data MsgDataGetSyncBlocks
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		//fmt.Println("msg.Decode(&data)")
		if pm.miner.State.CurrentHeight() < data.StartHeight {
			return nil // 不能提供
		}
		//fmt.Println("miner.GetGlobalInstanceHacashMiner()")
		// 查询区块并返回
		go func() {
			bkdb := store.GetGlobalInstanceBlocksDataStore()
			var blocks bytes.Buffer
			var blocklen = 0
			var blocksize = 0
			//fmt.Println("for height:=data.StartHeight; height <= minerdb.State.CurrentHeight(); height++ { ", data.StartHeight, pm.miner.State.CurrentHeight())
			for height := data.StartHeight; height <= pm.miner.State.CurrentHeight(); height++ {
				blkbytes, e := bkdb.GetBlockBytesByHeight(height, true, true)
				size := len(blkbytes)
				if e != nil || size == 0 {
					fmt.Printf(" not give block by height: %d, len: %d \n", height, size)
					return // 不能提供
				}
				//fmt.Printf("blocks = append(blocks, string(blkbytes)), height=%d, length=%d, string=%s \n", height, len(blkbytes), hex.EncodeToString(blkbytes))
				blocks.Write(blkbytes)
				blocksize += size
				blocklen++
				if blocklen >= 100 || blocksize >= 1024*1024 {
					//fmt.Println("totalsize >= 512KB break ")
					break
				}
			}
			if blocksize == 0 {
				//fmt.Println("blocks.Len() == 0")
				return
			}
			// 发送
			sdmsgdt := MsgDataSyncBlocks{
				FromHeight: data.StartHeight,
				ToHeight:   data.StartHeight + uint64(blocklen) - 1,
				Datas:      blocks.String(),
			}
			fmt.Printf("send sync blocks to peer %s, from height %d to %d \n", p.Name(), sdmsgdt.FromHeight, sdmsgdt.ToHeight)
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
			segblocks := make([]block.Block, 0, 100)
			segbodys := make([][]byte, 0, 100)
			stuffbytes := []byte(data.Datas)
			seek := uint32(0)
			for {
				if seek >= uint32(len(stuffbytes)) {
					break
				}
				blk, sk, e := blocks.ParseBlock(stuffbytes, seek)
				if e != nil {
					pm.syncMinerStatus(p) // 区块数据错误，重新同步
					return
				}
				segblocks = append(segblocks, blk)
				segbodys = append(segbodys, stuffbytes[seek:sk])
				seek = sk
			}
			fmt.Printf("got blocks (%d > %d), inserting ... ", data.FromHeight, data.ToHeight)
			insertCh := make(chan miner.DiscoveryNewBlockEvent, len(segblocks))
			subhandle := minerdb.SubscribeInsertBlock(insertCh)
			go func() { // 写入区块
				for i := 0; i < len(segblocks); i++ {
					//fmt.Println("minerdb.InsertBlock", segblocks[i].GetHeight())
					minerdb.InsertBlock(segblocks[i], segbodys[i])
				}
			}()
			for {
				insert := <-insertCh
				//fmt.Println("insert := <-insertCh ",insert.Block.GetHeight(), insert.Success)
				if !insert.Success {
					pm.removePeer(p.id) // 区块失败
					break
				}
				if insert.Block.GetHeight() == data.ToHeight { // insert ok
					subhandle.Unsubscribe()
					//fmt.Println("subhandle.Unsubscribe() pm.syncMinerStatus(p)")
					fmt.Printf("OK\n")
					pm.onsyncminer = false // 完成
					pm.syncMinerStatus(p)  // 再次同步
					break
				}
			}
		}()

	case msg.Code == NewBlockExcavateMsg:

		pm.Log.News("new block arrived to verify")
		// 新区块被挖出
		//fmt.Println("case msg.Code == NewBlockExcavateMsg: ")
		if pm.onsyncminer {
			pm.Log.Info("p2p on the syncing, not deal new block")
			return nil // 正在同步 忽略新挖出区块
		}
		pm.onsyncminer = true
		// 新区块被挖出
		//fmt.Println("var data MsgDataNewBlock")
		var data MsgDataNewBlock
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		//fmt.Println("err := msg.Decode(&data);", data.Height)
		// 如果区块高度超过，先同步
		mytarhei := pm.miner.State.CurrentHeight() + 1
		if mytarhei < data.Height {
			pm.Log.News("new block arrived height more than target, sync miner status right now")
			pm.syncMinerStatus(p)
			return nil
		}
		// 区块小于当前高度，忽略
		if mytarhei > data.Height {
			//fmt.Printf("mytarhei > data.Height %d>%d ...\n", mytarhei, data.Height)
			return nil
		}
		go func() {
			// 尝试添加区块
			blkbts := []byte(data.Datas)
			blk, _, e := blocks.ParseBlock(blkbts, 0)
			if e != nil {
				return
			}
			// 插入并等待
			insert := pm.miner.InsertBlockWait(blk, blkbts)
			if insert.Success { // insert ok
				str_time := time.Unix(int64(insert.Block.GetTimestamp()), 0).Format("01/02 15:04:05")
				fmt.Println("discovery new block, insert success.", "height", data.Height, "tx", len(insert.Block.GetTransactions())-1, "hash", hex.EncodeToString(insert.Block.Hash()), "prev", hex.EncodeToString(insert.Block.GetPrevHash()[0:16])+"...", "time", str_time)
				// 广播区块=
				data.block = insert.Block
				p.MarkBlock(insert.Block) // 标记区块
				go pm.BroadcastBlock(&data)
			}
			pm.onsyncminer = false // 状态恢复
			pm.miner.StartMining() // 可以开始挖矿

		}()

	case msg.Code == HeightHigherMsg:
		// 发现更高的区块高度，广播通知大家同步
		pm.Log.News("new block height to sync")
		if pm.onsyncminer {
			pm.Log.Info("p2p on the syncing, not deal higher height")
			return nil // 正在同步 忽略新高度
		}
		var data MsgDataHeightHigher
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg higher %v: %v", msg, err)
		}
		if data.Height <= pm.miner.State.CurrentHeight() {
			return nil // 区块高度并不太高
		}
		// 开始同步
		pm.onheighthigher = true
		pm.syncMinerStatus(p)

	default:
		return fmt.Errorf("%v", msg.Code)
	}

	return nil
}
