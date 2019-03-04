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
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
	"sync"
	"time"
)

type HacashMiner struct {
	State *MinerState

	TxPool service.TxPool

	CurrentActiveChainState *state.ChainState
	CurrentActiveCoinbase   block.Transaction

	// 可以开始挖矿
	canStartCh chan bool // 开始

	canPause bool

	// 正在插入区块
	insertBlock sync.Mutex

	// 成功挖掘新区快 事件订阅
	discoveryNewBlockFeed      event.Feed
	discoveryNewBlockFeedScope event.SubscriptionScope
}

type DiscoveryNewBlockEvent struct {
	Block block.Block
	Bodys []byte
}

var (
	globalInstanceHacashMinerMutex sync.Mutex
	globalInstanceHacashMiner      *HacashMiner = nil
)

func GetGlobalInstanceHacashMiner() *HacashMiner {
	globalInstanceHacashMinerMutex.Lock()
	defer globalInstanceHacashMinerMutex.Unlock()
	if globalInstanceHacashMiner == nil {
		globalInstanceHacashMiner = NewHacashMiner()
	}
	return globalInstanceHacashMiner
}

func NewHacashMiner() *HacashMiner {
	miner := &HacashMiner{}
	miner.State = NewMinerState()
	miner.State.DistLoad()
	miner.TxPool = txpool.GetGlobalInstanceMemTxPool()
	miner.canStartCh = make(chan bool, 1)
	miner.canPause = false
	return miner
}

// 通知可以开始挖矿
func (this *HacashMiner) CanStart() {
	this.canPause = false
	if len(this.canStartCh) == 0 {
		this.canStartCh <- true
	}
}

// 开始挖矿
func (this *HacashMiner) Start() error {
	go func() {
		for {
			this.start()
		}
	}()
	return nil
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
func (this *HacashMiner) CalculateTargetHash(block block.Block) ([]byte, uint32, error) {

	targetDifficulty := difficulty.CompactToBig(block.GetDifficulty())

	for i := uint32(0); i < uint32(4294967295); i++ {
		if this.canPause {
			return nil, 0, fmt.Errorf("canPause be set") // 暂停挖矿
		}
		// 休眠 秒
		time.Sleep(time.Duration(1200) * time.Microsecond)
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

// 生成一个新区块
func (this *HacashMiner) CreateBlock() (block.Block, error) {

	prev := this.State.prevBlockHead
	if prev == nil {
		prev = coin.GetGenesisBlock() // 创世
		this.State.prev288BlockTimestamp = uint64(time.Now().Unix())
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
	block, seek, e0 := blocks.ParseBlock(blockbytes, seek)
	if e0 != nil {
		return nil, 0, e0
	}
	//fmt.Printf("ArrivedNewBlockToUpdate(blockbytes []byte) block height: %d \n", block.GetHeight())
	// 验证签名
	sigok, e1 := block.VerifyNeedSigns()
	if e1 != nil {
		return nil, 0, e1
	}
	if !sigok {
		return nil, 0, fmt.Errorf("block signature verify faild")
	}
	// 锁定
	this.insertBlock.Lock()
	defer func() {
		fmt.Printf("ArrivedNewBlockToUpdate finish, height: %d \n", block.GetHeight())
		this.insertBlock.Unlock()
	}()

	//fmt.Printf("ArrivedNewBlockToUpdate  this.insertBlock.Lock()\n")

	// 判断当前状态
	if this.State.CurrentHeight()+1 != block.GetHeight() {
		return nil, 0, fmt.Errorf("not accepted block height")
	}
	if bytes.Compare(this.State.CurrentBlockHash(), block.GetPrevHash()) != 0 {
		return nil, 0, fmt.Errorf("not accepted block prev hash")
	}
	// 创建执行环境并执行
	newBlockChainState := state.NewTempChainState(nil)
	blksterr := block.ChangeChainState(newBlockChainState)
	if blksterr != nil {
		return nil, 0, blksterr
	}
	// 停止挖矿
	this.canPause = true
	// 保存区块，修改chain状态
	blockdb := store.GetGlobalInstanceBlocksDataStore()
	_, sverr := blockdb.Save(block)
	if sverr != nil {
		return nil, 0, sverr
	}
	chainstate := state.GetGlobalInstanceChainState()
	chainstate.TraversalCopy(newBlockChainState)
	// 修改矿工状态
	this.State.SetNewBlock(block)
	// 重新开始挖矿
	//this.Start()
	//this.canPause = false
	//this.CanStart()
	// 成功
	return block, seek, nil

}

// 订阅交易池加入新交易事件
func (this *HacashMiner) SubscribeDiscoveryNewBlock(discoveryCh chan<- DiscoveryNewBlockEvent) event.Subscription {
	return this.discoveryNewBlockFeedScope.Track(this.discoveryNewBlockFeed.Subscribe(discoveryCh))
}
