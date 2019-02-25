package miner

import (
	"fmt"
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
	"time"
)

type HacashMiner struct {
	State *MinerState

	TxPool service.TxPool

	CurrentActiveChainState *state.ChainState
	CurrentActiveCoinbase   block.Transaction
}

func NewHacashMiner() *HacashMiner {
	miner := &HacashMiner{}
	miner.State = NewMinerState()
	miner.State.DistLoad()
	miner.TxPool = txpool.GetGlobalInstanceMemTxPool()
	return miner
}

// 开始挖矿
func (this *HacashMiner) Start() error {

	for true {

		targetBlk, e := this.CreateBlock()
		if e != nil {
			return e
		}
		_, e1 := this.CalculateTargetHash(targetBlk)
		if e1 != nil {
			return e1
		}
		// 已挖到，保存状态数据
		store.GetBlocksDataStoreInstance().Save(targetBlk)
		state.GetGlobalInstanceChainState().TraversalCopy(this.CurrentActiveChainState)
		this.CurrentActiveChainState.Destroy()
		// 修改矿工状态信息，开始下一个区块挖矿
		this.State.prevBlockHead = targetBlk
		this.State.FlushSave()
		// ok one
		//fmt.Println( hex.EncodeToString(tarhash) )
	}

	return nil
}

// 生成一个新区块
func (this *HacashMiner) CalculateTargetHash(block block.Block) ([]byte, error) {

	for i := uint32(0); i < uint32(4294967295); i++ {
		//fmt.Println(i)
		block.SetNonce(i)
		tarhash := block.HashFresh()
		//fmt.Println( hex.EncodeToString(tarhash) )
		difficulty := difficulty.BigToCompact(difficulty.HashToBig(&tarhash))
		//fmt.Println(difficulty, block.GetDifficulty())
		if difficulty <= block.GetDifficulty() {
			// OK !!!!!!!!!!!!!!!
			return tarhash, nil
		}
		// 休眠 0.777 秒
		time.Sleep(time.Duration(777) * time.Millisecond)
	}
	// 尝试次数已达上限， 切换 coinbase
	this.CurrentActiveCoinbase.RecoverChainState(this.CurrentActiveChainState)
	this.SwitchBlockMinerAddress(block) // 切换
	this.CurrentActiveCoinbase.ChangeChainState(this.CurrentActiveChainState)
	return this.CalculateTargetHash(block)
	// 尝试次数已达上限
	//return nil, fmt.Errorf("CalculateTargetHash Attempts to reach the upper limit")
}

// 生成一个新区块
func (this *HacashMiner) CreateBlock() (block.Block, error) {

	prev := this.State.prevBlockHead
	if prev == nil {
		prev = coin.GetGenesisBlock() // 创世
	}
	nextblock := blocks.NewEmptyBlock_v1(prev)
	// 最低难度
	if nextblock.Difficulty == 0 {
		nextblock.Difficulty = fields.VarInt4(difficulty.LowestCompact)
	}
	// coinbase 占位
	nextblock.TransactionCount = 1
	nextblock.Transactions = append(nextblock.Transactions, nil)
	this.CurrentActiveChainState = state.NewTempChainState(nil)
	// 添加交易
	blockSize := uint32(block1def.ByteSizeBlockBeforeTransaction)
	for true {
		trs := this.TxPool.PopTxByHighestFee()
		if trs == nil {
			break
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
		this.CurrentActiveChainState.TraversalCopy(hxstate)
		hxstate.Destroy()
		nextblock.Transactions = append(nextblock.Transactions, trs)
		nextblock.TransactionCount += 1
	}
	// coinbase
	this.SwitchBlockMinerAddress(nextblock)
	this.CurrentActiveCoinbase.ChangeChainState(this.CurrentActiveChainState)
	// ok
	return nextblock, nil
}

func (this *HacashMiner) SwitchBlockMinerAddress(block block.Block) (block.Transaction, error) {

	// coinbase
	addrreadble := config.GetRandomMinerRewardAddress()
	addr, e := fields.CheckReadableAddress(addrreadble)
	if e != nil {
		return nil, e
	}
	coinbase := transactions.NewTransaction_0_Coinbase()
	coinbase.Address = *addr
	coinbase.Reward = *(coin.BlockCoinBaseReward(uint64(block.GetHeight())))
	block.GetTransactions()[0] = coinbase
	// 默克尔树
	root := blocks.CalculateMrklRoot(block.GetTransactions())
	block.SetMrklRoot(root)
	fmt.Println("miner coinbase: " + addrreadble + ", reward:" + coinbase.Reward.ToFinString())
	//
	this.CurrentActiveCoinbase = coinbase
	// ok
	return coinbase, nil
}
