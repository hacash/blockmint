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
	"sync/atomic"
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

type hashssync struct {
	p   *peer
	msg *MsgDataSyncHashs
}

type blockssync struct {
	p   *peer
	msg *MsgDataSyncBlocks
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
	txsyncCh     chan *txsync
	minersyncCh  chan *minersync
	heightsyncCh chan *heightsync

	// 状态同步
	hashssyncCh  chan *hashssync
	blockssyncCh chan *blockssync

	currentStatus uint32 // 当前的状态

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
		maxPeers: 25,
		peers:    newPeerSet(),
		txpool:   txpool.GetGlobalInstanceMemTxPool(),
		miner:    miner.GetGlobalInstanceHacashMiner(),
		Log:      log,
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

func (pm *ProtocolManager) GetPeers() *peerSet {
	return pm.peers
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

	// 状态同步
	pm.hashssyncCh = make(chan *hashssync)
	pm.blockssyncCh = make(chan *blockssync)

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
		pm.Log.Error("only one peer removed", peer.Name(), "no connect now, stop mining !")
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
	//var txs = make([]block.Transaction, 0, 16)
	// 读取交易
	txs := pm.txpool.GetTxs()
	if len(txs) > 0 {
		pm.Log.Debug("push txs to peer", p.Name(), "num", len(txs))
	}
	// 推给对方
	pm.txsyncCh <- &txsync{p, txs}
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncMinerStatus(p *peer) {
	pm.minersyncCh <- &minersync{p}
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) DoSyncMinerStatus(p *peer) {
	pm.Log.Noise("p2p sync miner status")
	if atomic.LoadUint32(&pm.currentStatus) > 0 {
		return // 正在同步状态，返回
	}
	peerone := pm.peers.Len() == 1
	blkhash, peer_height := p.Head()
	if bytes.Compare(pm.miner.State.CurrentBlockHash(), blkhash) == 0 {
		pm.Log.News("all block status is sync ok, peer head match")
		if peerone { // 开始挖矿
			pm.Log.Note("start mining...")
			pm.miner.StartMining()
		} else {

		}
		return // 状态一致，不需要同步
	}
	self_height := pm.miner.State.CurrentHeight()
	if peer_height <= self_height {
		pm.Log.News("all block status is sync ok, peer height less than or equal with me")
		if peerone { // 开始挖矿
			pm.Log.Note("start mining...")
			pm.miner.StartMining()
		}
		return // 高度小于我，或各自在一条高度相同的分叉上，不需要同步，等待下一个区块的出现
	}
	// 对方区块高度大于我，开始同步
	pm.miner.StopMining() // 停止挖矿
	pm.Log.Note("peer", p.Name(), "height", peer_height, "more than me height", self_height, "to check block fork and sync blocks")

	/* TODO: 同步区块，暂时关闭分叉检测 */
	match_height, err := pm.checkBlockFork(p, peer_height, self_height)
	if err != nil {
		pm.regainStatusToSimple() // 恢复状态
		pm.miner.StartMining()    // 开始挖矿
		pm.Log.Error("check block fork error:", err)
		return
	}
	var backblks []block.Block
	if match_height < self_height {
		// 回退区块
		pm.Log.News("back chain block height from", self_height, "to", match_height)
		backblks, err = pm.miner.BackTheWorldToHeight(match_height)
		if err != nil {
			pm.Log.Error("check block fork back to height", match_height, "error:", err)
		}
	}

	// 同步区块
	err = pm.syncBlocksFormPeer(p, match_height+1, peer_height)
	if err != nil {
		pm.Log.Error("check block fork sync blocks form peer", p.Name(), "error:", err)
		// 重新插入 恢复区块
		pm.miner.BackTheWorldToHeight(match_height)
		pm.miner.InsertBlocks(backblks)
	}
	// 同步区块完成，开始挖矿
	pm.regainStatusToSimple() // 恢复状态
	pm.miner.StartMining()

}

// 恢复状态
func (pm *ProtocolManager) regainStatusToSimple() {
	atomic.StoreUint32(&pm.currentStatus, SimpleStatus) // 恢复
}

// 读取同步区块
func (pm *ProtocolManager) syncBlocksFormPeer(p *peer, startHeight uint64, peer_height uint64) error {
	// 设置状态
	atomic.StoreUint32(&pm.currentStatus, SyncBlocksStatus) // 同步区块

	for {
		p2p.Send(p.rw, GetSyncBlocksMsg, MsgDataGetSyncBlocks{
			startHeight,
		})
		pm.Log.Debug("send GetSyncBlocksMsg to peer", p.Name(), "wait for data response")
		msg := <-pm.blockssyncCh // 等待读取消息
		if msg.p.id != p.id {
			return fmt.Errorf("sync blocks form peer %s not send msg to peer", msg.p.Name(), p.Name())
		}
		pm.Log.Info("got blocks height [", msg.msg.FromHeight, ",", msg.msg.ToHeight, "] insert to chain")
		data := msg.msg
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
				return fmt.Errorf("sync blocks parse block error:", e)
			}
			segblocks = append(segblocks, blk)
			segbodys = append(segbodys, stuffbytes[seek:sk])
			seek = sk
		}
		pm.Log.NoteEx(fmt.Sprintf("got blocks [%d,%d], inserting ... ", data.FromHeight, data.ToHeight))
		insertCh := make(chan miner.DiscoveryNewBlockEvent, len(segblocks))
		subhandle := pm.miner.SubscribeInsertBlock(insertCh)
		go func() { // 写入区块
			for i := 0; i < len(segblocks); i++ {
				//fmt.Println("minerdb.InsertBlock", segblocks[i].GetHeight())
				pm.miner.InsertBlock(segblocks[i], segbodys[i])
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
				pm.Log.Note("OK")
				break
			}
		}
		// 判断是否已经全部完成
		if data.ToHeight == peer_height {
			pm.Log.Note(fmt.Sprintf("all blocks height (%d,%d) sync finish", startHeight, peer_height))
			return nil // 全部区块同步完成 ！！！！
		}
		// 下一轮次
		startHeight = data.ToHeight + 1
	}

}

// 检查区块分叉
func (pm *ProtocolManager) checkBlockFork(p *peer, peer_height uint64, self_height uint64) (uint64, error) {
	if self_height == 0 {
		return 0, nil // 无区块
	}
	// 设置状态
	atomic.StoreUint32(&pm.currentStatus, SyncHashsStatus) // 正在检查分叉
	// 发起读取 区块哈希列表 hash
	startHeight := self_height
	endHight := self_height
	blkdb := store.GetGlobalInstanceBlocksDataStore()
	for {
		p2p.Send(p.rw, GetSyncHashsMsg, MsgDataGetSyncHashs{
			startHeight,
			endHight,
		})
		pm.Log.Debug("send GetSyncHashsMsg to peer", p.Name(), "wait for data response")
		msg := <-pm.hashssyncCh // 等待读取消息
		if msg.p.id != p.id {
			return 0, fmt.Errorf("check block fork get data from peer %s not send msg to peer %s", msg.p.Name(), p.Name())
		}
		// 检查hash是否匹配
		if msg.msg.StartHeight != startHeight || msg.msg.EndHeight != endHight {
			return 0, fmt.Errorf("check block fork peer %s send wrong hashs", p.Name())
		}
		for i := uint64(len(msg.msg.Hashs)) - 1; i > 0; i-- {
			lihash := []byte(msg.msg.Hashs[i])
			liheight := startHeight + i
			var selfheihash []byte
			if liheight == self_height {
				selfheihash = pm.miner.State.CurrentBlockHash()
			} else {
				var err error
				selfheihash, err = blkdb.GetBlockHashByHeight(liheight)
				if err != nil {
					return 0, fmt.Errorf("check block fork get self block hash error: %s", err)
				}
			}
			// 对比hash是否一致
			if bytes.Compare(lihash, selfheihash) == 0 {
				// 检查 哈希成功
				return liheight, nil // SUCCESS
			}
		}
		if endHight+100 < self_height {
			// 分叉太长，不能回退
			return 0, fmt.Errorf("block fork to long, cannot to switch back !!!")
		}
		// 都不匹配，则读取下一个片段
		endHight = startHeight - 1
		startHeight = endHight - 11
		if startHeight < 1 {
			return 0, fmt.Errorf("block fork to long, cannot to switch back !!!")
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
	defer func() {
		pm.Log.Info("peer", p.Name(), "be removed by handle msg end")
		pm.removePeer(p.id)
	}()

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
			pm.Log.Warning("peer", p.Name(), "handle msg error:", err)
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
		if len(sdtxs) > 0 {
			pm.Log.Info("txs be got num", len(sdtxs), "from peer", p.Name())
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
				pm.Log.Attention("add tx to pool error", pe)
			} else {
				pm.Log.News("add tx", hex.EncodeToString(tx.Hash()), "to my txpool success")
			}
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
					fmt.Printf(" not give block by height: %d, len: %d, error: %s\n", height, size, e)
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
		// 检查状态
		if atomic.LoadUint32(&pm.currentStatus) != SyncBlocksStatus {
			return nil
		}
		// 抛出数据
		pm.blockssyncCh <- &blockssync{
			p,
			&data,
		}

		/*

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

		*/

	case msg.Code == GetSyncHashsMsg:
		// 获取hash列表
		var data MsgDataGetSyncHashs
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		self_height := pm.miner.State.CurrentHeight()
		if self_height < data.EndHeight {
			return nil // 不能提供
		}
		// 查询hash并返回
		go func() {
			bkdb := store.GetGlobalInstanceBlocksDataStore()
			var hashs = make([]string, 0)
			for height := data.StartHeight; height <= self_height && height <= data.EndHeight; height++ {
				hash, e := bkdb.GetBlockHashByHeight(height)
				if e != nil {
					return
				}
				//fmt.Println("append string(hash)", hash)
				hashs = append(hashs, string(hash))
			}
			// 发送
			sdmsgdt := MsgDataSyncHashs{
				StartHeight: data.StartHeight,
				EndHeight:   data.EndHeight,
				Hashs:       hashs,
			}
			pm.Log.Info(fmt.Sprintf("send sync hashs to peer %s height <%d,%d>", p.Name(), sdmsgdt.StartHeight, sdmsgdt.EndHeight))
			go p2p.Send(p.rw, SyncHashsMsg, sdmsgdt)
		}()

	case msg.Code == SyncHashsMsg:
		//
		var data MsgDataSyncHashs
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg %v: %v", msg, err)
		}
		// 检查状态
		if atomic.LoadUint32(&pm.currentStatus) != SyncHashsStatus {
			return nil
		}
		// 抛出数据
		pm.hashssyncCh <- &hashssync{
			p,
			&data,
		}

	case msg.Code == NewBlockExcavateMsg:
		// 检查状态
		if atomic.LoadUint32(&pm.currentStatus) != SimpleStatus {
			pm.Log.Debug("p2p on the syncing not to deal NewBlockExcavateMsg")
			return nil
		}
		pm.Log.News("new block arrived to verify")
		// 新区块被挖出
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
				pm.Log.Note("discovery new block, insert success.", "height", data.Height, "tx", len(insert.Block.GetTransactions())-1, "hash", hex.EncodeToString(insert.Block.Hash()), "prev", hex.EncodeToString(insert.Block.GetPrevHash()[0:16])+"...", "time", str_time)
				// 广播区块=
				data.block = insert.Block
				p.MarkBlock(insert.Block) // 标记区块
				go pm.BroadcastBlock(&data)
			}
			// 状态恢复
			pm.miner.StartMining() // 可以开始挖矿

		}()

	case msg.Code == HeightHigherMsg:
		// 发现更高的区块高度，广播通知大家同步
		pm.Log.News("new block height to sync")
		var data MsgDataHeightHigher
		if err := msg.Decode(&data); err != nil {
			return fmt.Errorf("msg higher %v: %v", msg, err)
		}
		if data.Height <= pm.miner.State.CurrentHeight() {
			return nil // 区块高度并不太高
		}
		// 开始同步
		pm.syncMinerStatus(p)

	default:
		return fmt.Errorf("%v", msg.Code)
	}

	return nil
}
