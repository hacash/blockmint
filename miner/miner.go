package miner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
	"strconv"
	"sync"
	"time"
)

var (
	insertBlocksChSize = 255

	miningSleepNanosecond = uint64(0) // 矿工休眠时间

)

type HacashMiner struct {
	State *MinerState

	TxPool service.TxPool

	// 矿工状态标识
	miningStatusCh chan bool

	// 正在插入区块
	insertBlock sync.Mutex

	// 当前能否添加区块
	insertBlocksCh chan *DiscoveryNewBlockEvent

	// 插入区块进度 事件订阅
	insertBlockFeed      event.Feed
	insertBlockFeedScope event.SubscriptionScope

	// 成功挖掘新区快 事件订阅
	discoveryNewBlockFeed      event.Feed
	discoveryNewBlockFeedScope event.SubscriptionScope

	Log log.Logger
}

type DiscoveryNewBlockEvent struct {
	Success bool // 成功写入
	Block   block.Block
	Bodys   []byte
}

var (
	globalInstanceHacashMinerMutex sync.Mutex
	globalInstanceHacashMiner      *HacashMiner = nil
)

func GetGlobalInstanceHacashMiner() *HacashMiner {
	globalInstanceHacashMinerMutex.Lock()
	defer globalInstanceHacashMinerMutex.Unlock()
	if globalInstanceHacashMiner == nil {
		lg := config.GetGlobalInstanceLogger()
		globalInstanceHacashMiner = NewHacashMiner(lg)
	}
	return globalInstanceHacashMiner
}

func NewHacashMiner(logger log.Logger) *HacashMiner {
	// 检查配置
	sleepnano, e1 := strconv.ParseUint(config.Config.Miner.Stepsleepnano, 10, 0)
	if e1 != nil {
		panic("config.Config.Miner.Stepsleepnano not be " + config.Config.Miner.Stepsleepnano)
	}
	if sleepnano > 0 {
		logger.Note("miner step calculation sleep", sleepnano, "nanosecond")
	}
	miningSleepNanosecond = sleepnano
	// 创建
	miner := &HacashMiner{}
	miner.Log = logger
	miner.State = NewMinerState(logger)
	miner.State.FetchLoad()
	miner.TxPool = txpool.GetGlobalInstanceMemTxPool()
	miner.miningStatusCh = make(chan bool, 200)

	miner.insertBlocksCh = make(chan *DiscoveryNewBlockEvent, insertBlocksChSize)

	return miner
}

// 开始挖矿
func (this *HacashMiner) Start() {
	go this.miningLoop()
	go this.insertBlockLoop()
}

// 开始挖矿
func (this *HacashMiner) StartMining() {
	//this.Log.Noise("hacash miner will start mining by call func StartMining()")
	this.miningStatusCh <- true

	/*
		if 0 == len(this.startingCh) {
			// 如果是停止状态
			this.Log.Info("start mining")
			this.startingCh <- true
			// 切换到
		}
	*/
}

// 开始挖矿
func (this *HacashMiner) StopMining() {
	//this.Log.Noise("hacash miner will stop mining by call func StopMining()")
	this.miningStatusCh <- false
	/*
		if 1 == atomic.LoadUint32(&this.miningStatus) {
			this.Log.Info("stop mining")
			this.stopingCh <- true
			//this.Log.Noise("stop mining ok!!!")
		}
	*/
}

// 挖矿循环
func (this *HacashMiner) miningLoop() {
	for {
		this.Log.Info("mining loop wait to start")
		select {
		case stat := <-this.miningStatusCh:
			if stat == false {
				continue // 停止状态
			}
			this.Log.Info("do mining start")
			err := this.doMining()
			if err != nil {
				this.Log.Info("mining process out for", err)
			} else {
				this.Log.Info("mining process out")
				// 继续挖掘下一个区块
				this.StartMining()
			}
		}
	}
}

// 执行挖矿
func (this *HacashMiner) doMining() error {
	// 创建区块
	newBlock, _, coinbase, _, e := this.CreateNewBlock()
	if e != nil {
		this.Log.Warning("create new block for mining error", e)
		return e
	}
	// 挖掘计算
	var targetHash []byte
	targetDifficulty := newBlock.GetDifficulty()
RESTART_TO_MINING:
	rewardAddrReadble := this.setMinerForCoinbase(coinbase)                    // coinbase
	newBlock.SetMrklRoot(blocks.CalculateMrklRoot(newBlock.GetTransactions())) // update mrkl root
	this.Log.Note("set new coinbase address", rewardAddrReadble, "height", newBlock.GetHeight(), "do mining...")
	for i := uint32(0); i < 4294967295; i++ {
		// this.Log.Noise(i)
		select {
		case stat := <-this.miningStatusCh:
			if stat == false {
				this.reputAllTxsFromBlock(newBlock) // 重新放入所有交易到交易池
				// this.Log.Debug("mining break and stop mining -…………………………………………………………………………")
				return fmt.Errorf("mining break by set sign stoping chan") // 停止挖矿
			}
		default:

		}
		if miningSleepNanosecond > 0 {
			time.Sleep(time.Duration(miningSleepNanosecond) * time.Nanosecond)
		}
		newBlock.SetNonce(i)
		targetHash = newBlock.HashFresh()
		curdiff := difficulty.BigToCompact(difficulty.HashToBig(&targetHash))
		//fmt.Println(curdiff, targetDifficulty)
		if curdiff < targetDifficulty {
			this.Log.Info("find a valid nonce for block", "height", newBlock.GetHeight())
			// OK !!!!!!!!!!!!!!!
			goto MINING_SUCCESS
		}
	}
	goto RESTART_TO_MINING // 下一轮次
MINING_SUCCESS:

	// 插入并等待结果
	insert := this.InsertBlockWait(newBlock, nil)
	if insert.Success {
		targethashhex := hex.EncodeToString(targetHash)
		// 广播新区快信息
		this.Log.Info("mining success one block", "hash", targethashhex)
		go this.discoveryNewBlockFeed.Send(DiscoveryNewBlockEvent{
			Block: newBlock,
			Bodys: insert.Bodys,
		})
		// 打印相关信息
		str_time := time.Unix(int64(newBlock.GetTimestamp()), 0).Format("01/02 15:04:05")
		this.Log.Note(fmt.Sprintf("bh: %d, tx: %d, df: %d, hx: %s, px: %s, cm: %s, rw: %s, tt: %s",
			int(newBlock.GetHeight()),
			len(newBlock.GetTransactions())-1,
			newBlock.GetDifficulty(),
			targethashhex,
			hex.EncodeToString(newBlock.GetPrevHash()[0:16])+"...",
			rewardAddrReadble,
			coinbase.Reward.ToFinString(),
			str_time,
		))
	} else {
		length := this.reputAllTxsFromBlock(newBlock)
		this.Log.Warning("mining finish block", "height", newBlock.GetHeight(), "hash", hex.EncodeToString(newBlock.Hash()), "insert chain fail, reappend all txs", length, "and clear block")
	}
	return nil
}

// 将交易全部放入交易池
func (this *HacashMiner) reputAllTxsFromBlock(newBlock block.Block) int {
	blktxs := newBlock.GetTransactions()
	length := len(blktxs) - 1
	for i := length; i > 0; i-- { // drop coinbase
		this.TxPool.AddTx(blktxs[i]) // 倒序重新放入交易池
	}
	return length
}

// 插入区块
func (this *HacashMiner) InsertBlock(blk block.Block, bodys []byte) {
	this.insertBlocksCh <- &DiscoveryNewBlockEvent{
		false,
		blk,
		bodys,
	}
}

// 插入区块
func (this *HacashMiner) InsertBlocks(blks []block.Block) {
	go func() {

		for _, blk := range blks {
			this.insertBlocksCh <- &DiscoveryNewBlockEvent{
				false,
				blk,
				nil,
			}
		}

	}()
}

// 插入区块
func (this *HacashMiner) insertBlockLoop() {
	for {
		//tk := time.NewTimer(time.Second * 9)
		select {
		case blk := <-this.insertBlocksCh:
			this.Log.Info("insert loop get one block", "height", blk.Block.GetHeight())
			// tk.Stop()
			this.StopMining() // 停止挖矿
			err := this.doInsertBlock(blk)
			if err != nil {
				this.Log.Info("insert block loop", "height", blk.Block.GetHeight(), "error", err)
			} else {
				this.Log.Info("insert block ok", "height", blk.Block.GetHeight())
			}
			//case <-tk.C:
			// this.Log.Noise("no block to insert")
			// this.StartMining() // 几秒后没有区块插入则自动开始挖矿
		}
	}
}

// 插入区块
func (this *HacashMiner) doInsertBlock(blk *DiscoveryNewBlockEvent) error {
	if blk.Block == nil && (blk.Bodys == nil || len(blk.Bodys) == 0) {
		fmt.Errorf("data is empty")
	}
	if blk.Block == nil {
		b, _, e := blocks.ParseBlock(blk.Bodys, 0)
		if e != nil {
			return e
		}
		blk.Block = b
	}
	block := blk.Block
	successInsert := false
	blockBytes := []byte{}
	defer func() {
		// 插入处理事件通知
		go this.insertBlockFeed.Send(DiscoveryNewBlockEvent{
			successInsert,
			block,
			blockBytes,
		})
	}()
	// 判断高度
	var fail_height = this.State.CurrentHeight()+1 != block.GetHeight()
	var fail_prevhash = bytes.Compare(this.State.CurrentBlockHash(), block.GetPrevHash()) != 0
	if fail_height || fail_prevhash {
		var typestr = "prev hash"
		if fail_height {
			typestr += "height"
		}
		return fmt.Errorf("not accepted block with wrong %s, height=%d, hash=%s, target_prev_hash=%s, base_height=%d, base_hash=%s, base_prev_hash=%s",
			typestr,
			block.GetHeight(),
			hex.EncodeToString(block.Hash()),
			hex.EncodeToString(block.GetPrevHash()),
			this.State.CurrentHeight(),
			hex.EncodeToString(this.State.CurrentBlockHash()),
			hex.EncodeToString(this.State.prevBlockHead.GetPrevHash()),
		)
	}
	// 检查难度值
	blkhash := block.HashFresh()
	hxdift := difficulty.BigToCompact(difficulty.HashToBig(&blkhash))
	tardift := this.State.TargetDifficultyCompact(block.GetHeight(), nil)
	if hxdift > tardift {
		return fmt.Errorf("difficulty not satisfy, height %d, accept %d, but got %d", block.GetHeight(), tardift, hxdift)
	}
	// 验证签名
	sigok, e1 := block.VerifyNeedSigns()
	if e1 != nil {
		return e1
	}
	if !sigok {
		return fmt.Errorf("block signature verify faild")
	}
	// 判断交易是否已经存在
	blockdb := store.GetGlobalInstanceBlocksDataStore()
	txs := block.GetTransactions()
	if len(txs) < 1 {
		return fmt.Errorf("block is empty")
	}
	for i := 1; i < len(txs); i++ {
		txhashnofee := txs[i].HashNoFee()
		if ok, e := blockdb.CheckTransactionExist(txhashnofee); ok || e != nil {
			return fmt.Errorf("tx %s is exist", hex.EncodeToString(txhashnofee))
		}
	}
	// 验证交易
	newBlockChainState := state.NewTempChainState(nil)
	blksterr := block.ChangeChainState(newBlockChainState)
	if blksterr != nil {
		return blksterr
	}
	// 存储区块数据
	var sverr error
	blockBytes, sverr = blockdb.Save(block)
	if sverr != nil {
		return sverr
	}
	// 保存用户状态
	chainstate := state.GetGlobalInstanceChainState()
	chainstate.TraversalCopy(newBlockChainState)
	// 从交易池内清除已经出块的交易
	this.TxPool.RemoveTxs(block.GetTransactions())
	// 更新矿工状态
	this.State.SetNewBlock(block)
	// 成功状态
	successInsert = true

	return nil
}

// 插入区块进度事件
func (this *HacashMiner) SubscribeInsertBlock(insertCh chan<- DiscoveryNewBlockEvent) event.Subscription {
	return this.insertBlockFeedScope.Track(this.insertBlockFeed.Subscribe(insertCh))
}

// 订阅挖掘出新区快事件
func (this *HacashMiner) SubscribeDiscoveryNewBlock(discoveryCh chan<- DiscoveryNewBlockEvent) event.Subscription {
	return this.discoveryNewBlockFeedScope.Track(this.discoveryNewBlockFeed.Subscribe(discoveryCh))
}

// 插入区块，等待插入状态返回
func (this *HacashMiner) InsertBlockWait(blk block.Block, bodys []byte) DiscoveryNewBlockEvent {
	insertCh := make(chan DiscoveryNewBlockEvent, 1)
	subhandle := this.SubscribeInsertBlock(insertCh)
	// 写入区块
	this.Log.Debug("insert block to chain state with wait", "height", blk.GetHeight())
	this.InsertBlock(blk, bodys)
	// 等待返回
	for {
		res := <-insertCh
		if bytes.Compare(res.Block.Hash(), blk.Hash()) == 0 {
			this.Log.Debug("insert block return", "height", blk.GetHeight())
			subhandle.Unsubscribe() // 取消注册
			return res
		}
	}
}

// 创建区块
func (this *HacashMiner) CreateNewBlock() (block.Block, *state.ChainState, *transactions.Transaction_0_Coinbase, *fields.Amount, error) {
	nextblock := blocks.NewEmptyBlock_v1(this.State.prevBlockHead)

	//////////////// test ////////////////
	//timeset := coin.GetGenesisBlock().GetTimestamp() + nextblock.GetHeight()*300 - 150 + uint64(rand.Int63n(300))
	//nextblock.Timestamp = fields.VarInt5(timeset)
	//if timeset > uint64(time.Now().Unix()) {
	//	panic("okokokok")
	//}
	//////////////// test ////////////////

	hei, dfct, info := this.State.NextHeightTargetDifficultyCompact()
	if info != nil && *info != "" {
		this.Log.Note(*info)
	}
	nextblock.Height = fields.VarInt5(hei)
	nextblock.Difficulty = fields.VarInt4(dfct)
	coinbase := this.createCoinbaseTx(nextblock)
	nextblock.TransactionCount = 1
	nextblock.Transactions = append(nextblock.Transactions, coinbase)
	// 获取交易并验证
	tempBlockState := state.NewTempChainState(nil)
	// 添加交易
	stoblk := store.GetGlobalInstanceBlocksDataStore()
	blockSize := uint32(block1def.ByteSizeBlockBeforeTransaction)
	blockTotalFee := fields.NewEmptyAmount()
	for {
		trs := this.TxPool.PopTxByHighestFee()
		if trs == nil {
			break // nothing
		}
		hashnofee := trs.HashNoFee() // 交易是否已经存在
		ext, e2 := stoblk.CheckTransactionExist(hashnofee)
		if e2 != nil || ext {
			continue // drop tx
		}
		blockSize += trs.Size()
		if int64(blockSize) > config.MaximumBlockSize {
			this.TxPool.AddTx(trs)
			break // over block size
		}
		hxstate := state.NewTempChainState(tempBlockState)
		errun := trs.ChangeChainState(hxstate)
		if errun != nil {
			hxstate.Destroy()
			continue // error , give up tx
		}
		// ok copy state
		tempBlockState.TraversalCopy(hxstate)
		hxstate.Destroy()
		nextblock.Transactions = append(nextblock.Transactions, trs)
		nextblock.TransactionCount += 1
		// 手续费
		fee := fields.ParseAmount(trs.GetFee(), 0)
		blockTotalFee, _ = blockTotalFee.Add(fee)
	}
	this.Log.Info("create new block", "height", nextblock.Height, "transaction", nextblock.TransactionCount-1)

	return nextblock, tempBlockState, coinbase, blockTotalFee, nil
}

// 创建coinbase交易
func (this *HacashMiner) createCoinbaseTx(block block.Block) *transactions.Transaction_0_Coinbase {
	// coinbase
	coinbase := transactions.NewTransaction_0_Coinbase()
	coinbase.Reward = *(coin.BlockCoinBaseReward(uint64(block.GetHeight())))
	this.setMinerForCoinbase(coinbase)
	return coinbase
}

// 设置coinbase交易
func (this *HacashMiner) setMinerForCoinbase(coinbase *transactions.Transaction_0_Coinbase) string {
	addrreadble := config.GetRandomMinerRewardAddress()
	addr, e := fields.CheckReadableAddress(addrreadble)
	if e != nil {
		panic("Miner Reward Address `" + addrreadble + "` Error !")
	}
	coinbase.Address = *addr
	return addrreadble
}

// 倒退区块
func (this *HacashMiner) BackTheWorldToHeight(target_height uint64) ([]block.Block, error) {

	current_height := this.State.CurrentHeight()
	if target_height >= current_height {
		// do nothing
		return []block.Block{}, nil
	}

	// 被回退的区块
	var backblks = make([]block.Block, current_height-target_height-1)

	db := store.GetGlobalInstanceBlocksDataStore()
	state := state.GetGlobalInstanceChainState()
	for {
		blkbts, err := db.GetBlockBytesByHeight(current_height, true, true)
		if err != nil {
			return backblks, err
		}
		blkobj, _, e2 := blocks.ParseBlock(blkbts, 0)
		if e2 != nil {
			return backblks, e2
		}
		backblks = append(backblks, blkobj)
		this.Log.Note("delete height", current_height, "hash", hex.EncodeToString(blkobj.Hash()), "prev_hash", hex.EncodeToString(blkobj.GetPrevHash()[0:16])+"...")
		// 回退状态
		blkobj.RecoverChainState(state)
		// 删除数据
		db.DeleteBlockForceUnsafe(blkobj)
		// 是否完成
		current_height--
		if current_height <= target_height {
			break
		}
	}
	// 修改矿工状态
	blkhdbts, e0 := db.GetBlockBytesByHeight(current_height, true, true)
	if e0 != nil {
		return backblks, e0
	}
	//fmt.Println("head bytes ", hex.EncodeToString(blkhdbts))
	blkhead, _, e2 := blocks.ParseBlock(blkhdbts, 0)
	if e2 != nil {
		return backblks, e2
	}
	prev288blkhei := current_height - (current_height % config.ChangeDifficultyBlockNumber)
	if prev288blkhei == 0 {
		rootblk := coin.GetGenesisBlock()
		this.State.prev288BlockTimestamp = rootblk.GetTimestamp() // 起始时间戳
	} else {
		blkhdbts_prev288, e1 := db.GetBlockBytesByHeight(prev288blkhei, true, false)
		if e1 != nil {
			return backblks, e1
		}
		blkhead_prev288, _, e3 := blocks.ParseBlockHead(blkhdbts_prev288, 0)
		if e3 != nil {
			return backblks, e3
		}
		this.State.prev288BlockTimestamp = blkhead_prev288.GetTimestamp() // 时间戳
		//fmt.Println("修改矿工状态 height", blkhead.GetHeight(), "hash", hex.EncodeToString(blkhead.Hash()))
	}
	this.State.SetNewBlock(blkhead)
	// ok
	return backblks, nil
}

/////////////////////////////////////////////////////////////////////////////////

/*



// 计算hash
func (this *HacashMiner) CalculateTargetHash(block block.Block) ([]byte, uint32, error) {

	targetDifficulty := difficulty.CompactToBig(block.GetDifficulty())

	for i := uint32(0); i < uint32(4294967295); i++ {
		if this.miningBreakOnce {
			return nil, 0, fmt.Errorf("miningBreakOnce be set") // 暂停挖矿
		}
		// 休眠 秒
		time.Sleep(time.Duration(10) * time.Millisecond)
		//fmt.Println(i)
		block.SetNonce(i)
		tarhash := block.HashFresh()
		//fmt.Println( hex.EncodeToString(tarhash) )
		curdiff := difficulty.HashToBig(&tarhash)
		//fmt.Println(difficulty, block.GetDifficulty())
		if curdiff.Cmp(targetDifficulty) == -1 {
			// OK !!!!!!!!!!!!!!!
			return tarhash, i, nil
		}
	}
	// 尝试次数已达上限， 切换 reward address
	this.SwitchBlockMinerAddress(block)
	return this.CalculateTargetHash(block)
}


// 等待外部信号，而后开始
func (this *HacashMiner) start() error {

	// 等待开始
	<-this.canStartCh

	totalTime := uint64(0)
	totalHashTry := uint64(0)
	for i := uint64(1); ; i++ {
		t1 := time.Now() // get current time
		targetBlk, e := this.CreateBlock()
		if e != nil {
			return e
		}
		tarhash, trynum, e1 := this.CalculateTargetHash(targetBlk)
		if e1 != nil {
			return e1
		}
		totalHashTry += uint64(trynum)
		// 锁定
		this.insertBlock.Lock()
		// 检查矿工状态
		if bytes.Compare(this.State.CurrentBlockHash(), targetBlk.GetPrevHash()) != 0 {
			return fmt.Errorf("miner state change") // 状态改变，挖下一个区块
		}
		// 已挖到，保存状态数据
		bodys, _ := store.GetGlobalInstanceBlocksDataStore().Save(targetBlk)
		state.GetGlobalInstanceChainState().TraversalCopy(this.CurrentActiveChainState)
		this.CurrentActiveChainState.Destroy()
		this.CurrentActiveChainState = nil
		// 修改矿工状态信息，开始下一个区块挖矿
		this.State.SetNewBlock(targetBlk)
		// 解锁
		this.insertBlock.Unlock()
		// 发送挖掘出新区块消息
		go this.discoveryNewBlockFeed.Send(DiscoveryNewBlockEvent{
			Block: targetBlk,
			Bodys: bodys,
		})
		// ok one
		//fmt.Println( hex.EncodeToString(tarhash) )
		elapsed := time.Since(t1)
		spt := int(elapsed.Seconds())
		totalTime += uint64(spt)
		fmt.Printf("bh: %d, hs: %s, cn: %d, ts: %ds, ah: %d, at: %ds \n",
			int(targetBlk.GetHeight()),
			hex.EncodeToString(tarhash),
			trynum,
			spt,
			totalHashTry/i,
			totalTime/i,
		)

	}

	return nil
}

// 生成一个新区块
func (this *HacashMiner) createBlock() (block.Block, error) {

	prev := this.State.prevBlockHead
	if prev == nil {
		prev = coin.GetGenesisBlock()                          // 创世
		this.State.prev288BlockTimestamp = prev.GetTimestamp() // uint64(time.Now().Unix())
	}
	nextblock := blocks.NewEmptyBlock_v1(prev)
	// 最低难度
	if uint64(nextblock.Height) < uint64(config.ChangeDifficultyBlockNumber) {
		nextblock.Difficulty = fields.VarInt4(difficulty.LowestCompact)
	} else {
		newDifficulty := difficulty.CalculateNextWorkTarget(
			uint32(nextblock.Difficulty),
			uint64(nextblock.Height),
			this.State.prev288BlockTimestamp,
			uint64(nextblock.Timestamp),
		)
		if newDifficulty != uint32(nextblock.Difficulty) {
			//fmt.Printf("CalculateNextWorkTarget new Difficulty  ============================  %d \n", newDifficulty)
			this.State.prev288BlockTimestamp = uint64(nextblock.Timestamp)
			nextblock.Difficulty = fields.VarInt4(newDifficulty) // 计算难度
		}
	}
	// coinbase 占位
	nextblock.TransactionCount = 1
	nextblock.Transactions = append(nextblock.Transactions, nil)
	if this.CurrentActiveChainState != nil {
		this.CurrentActiveChainState.Destroy()
		this.CurrentActiveChainState = nil
	}
	this.CurrentActiveChainState = state.NewTempChainState(nil)
	// 添加交易
	stoblk := store.GetGlobalInstanceBlocksDataStore()
	blockSize := uint32(block1def.ByteSizeBlockBeforeTransaction)
	blockTotalFee := fields.NewEmptyAmount()
	for true {
		trs := this.TxPool.PopTxByHighestFee()
		if trs == nil {
			break
		}
		hashnofee := trs.HashNoFee()
		ext, e2 := stoblk.CheckTransactionExist(hashnofee)
		if e2 != nil || ext {
			break // trs already exist
		}
		blockSize += trs.Size()
		if int64(blockSize) > config.MaximumBlockSize {
			this.TxPool.AddTx(trs)
			break // over block size
		}
		hxstate := state.NewTempChainState(this.CurrentActiveChainState)
		errun := trs.ChangeChainState(hxstate)
		if errun != nil {
			hxstate.Destroy()
			break // error
		}
		// ok copy state
		//fmt.Println("ok copy hxstate state ======================")
		this.CurrentActiveChainState.TraversalCopy(hxstate)
		hxstate.Destroy()
		nextblock.Transactions = append(nextblock.Transactions, trs)
		nextblock.TransactionCount += 1
		var fee fields.Amount
		fee.Parse(trs.GetFee(), 0)
		blockTotalFee.Add(&fee)
	}
	// coinbase
	coinbase, _ := this.CreateBlockMinerAddress(nextblock, blockTotalFee)
	coinbase.ChangeChainState(this.CurrentActiveChainState) // 加上奖励和手续费
	// ok
	return nextblock, nil
}

func (this *HacashMiner) CreateBlockMinerAddress(block block.Block, totalFee *fields.Amount) (block.Transaction, error) {

	// coinbase
	addrreadble := config.GetRandomMinerRewardAddress()
	addr, e := fields.CheckReadableAddress(addrreadble)
	if e != nil {
		return nil, e
	}
	coinbase := transactions.NewTransaction_0_Coinbase()
	coinbase.Address = *addr
	coinbase.Reward = *(coin.BlockCoinBaseReward(uint64(block.GetHeight())))
	coinbase.TotalFee = *totalFee
	block.GetTransactions()[0] = coinbase
	// 默克尔树
	root := blocks.CalculateMrklRoot(block.GetTransactions())
	block.SetMrklRoot(root)
	timeNow := time.Now()
	fmt.Printf("bh: %d, tx: %d, tt: %s-%d %d:%d, df: %d, cm: %s, rw: %s \n",
		int(block.GetHeight()),
		len(block.GetTransactions())-1,
		timeNow.Month().String(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(),
		block.GetDifficulty(),
		// hex.EncodeToString(block.GetPrevHash()[0:16]) + "...",
		addrreadble,
		coinbase.Reward.ToFinString(),
	)
	//
	this.CurrentActiveCoinbase = coinbase
	// ok
	return coinbase, nil
}

func (this *HacashMiner) SwitchBlockMinerAddress(block block.Block) (block.Transaction, error) {
	if this.CurrentActiveCoinbase == nil {
		panic("CurrentActiveCoinbase is nil")
	}
	var totalFee fields.Amount
	totalFee.Parse(this.CurrentActiveCoinbase.GetFee(), 0)
	this.CurrentActiveCoinbase.RecoverChainState(this.CurrentActiveChainState) // 回退
	coinbase, e := this.CreateBlockMinerAddress(block, &totalFee)
	coinbase.ChangeChainState(this.CurrentActiveChainState) // 切换
	return coinbase, e
}

// 新到了一个区块，验证并写入，更新矿工状态
func (this *HacashMiner) ArrivedNewBlockToUpdate(blockbytes []byte, seek uint32) (block.Block, uint32, error) {
	fmt.Printf("ArrivedNewBlockToUpdate ================ \n")
	block, seek, e0 := blocks.ParseBlock(blockbytes, seek)
	if e0 != nil {
		return nil, 0, e0
	}
	fmt.Printf("ArrivedNewBlockToUpdate(blockbytes []byte) block height: %d \n", block.GetHeight())
	// 验证签名
	sigok, e1 := block.VerifyNeedSigns()
	if e1 != nil {
		fmt.Println("ArrivedNewBlockToUpdate  block signature verify error")
		return nil, 0, e1
	}
	if !sigok {
		fmt.Println("ArrivedNewBlockToUpdate  block signature verify faild")
		return nil, 0, fmt.Errorf("block signature verify faild")
	}
	fmt.Printf("ArrivedNewBlockToUpdate  this.insertBlock.Lock()\n")

	// 锁定
	//this.insertBlock.Lock()
	//defer func() {
	//	fmt.Printf("ArrivedNewBlockToUpdate finish, height: %d \n", block.GetHeight())
	//	this.insertBlock.Unlock()
	//}()

	fmt.Printf("ArrivedNewBlockToUpdate  this.State.CurrentHeight()+1 != block.GetHeight()\n")

	// 判断当前状态
	if this.State.CurrentHeight()+1 != block.GetHeight() {
		fmt.Printf("ArrivedNewBlockToUpdate  not accepted block height\n")
		return nil, 0, fmt.Errorf("not accepted block height")
	}
	if bytes.Compare(this.State.CurrentBlockHash(), block.GetPrevHash()) != 0 {
		fmt.Printf("ArrivedNewBlockToUpdate not accepted block prev hash\n")
		return nil, 0, fmt.Errorf("not accepted block prev hash")
	}
	// 创建执行环境并执行
	fmt.Printf("newBlockChainState := state.NewTempChainState(nil)\n")
	newBlockChainState := state.NewTempChainState(nil)
	blksterr := block.ChangeChainState(newBlockChainState)
	if blksterr != nil {
		fmt.Printf("blksterr := block.ChangeChainState(newBlockChainState  error: %s \n", blksterr)
		return nil, 0, blksterr
	}
	fmt.Printf("停止挖矿 this.miningBreakOnce = true\n")
	// 停止挖矿
	this.miningBreakOnce = true
	// 保存区块，修改chain状态


	blockdb := store.GetGlobalInstanceBlocksDataStore()
	_, sverr := blockdb.Save(block)
	if sverr != nil {
		fmt.Printf(" _, sverr := blockdb.Save(block)  error: %s \n", sverr)
		return nil, 0, sverr
	}
	fmt.Printf("chainstate.TraversalCopy(newBlockChainState)\n")
	chainstate := state.GetGlobalInstanceChainState()
	chainstate.TraversalCopy(newBlockChainState)
	// 修改矿工状态
	this.State.SetNewBlock(block)



	// 重新开始挖矿
	//this.Start()
	//this.miningBreakOnce = false
	//this.CanStart()
	// 成功
	fmt.Printf("成功 return block, seek, nil\n")
	return block, seek, nil

}

*/
