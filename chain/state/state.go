package state

import (
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/miner"
	"math/rand"
	"os"
	"path"
	"strconv"
	"sync"
)

//
type ChainState struct {
	tempdir string

	balanceDB *db.BalanceDB
	diamondDB *db.DiamondDB

	// 正在处理的区块
	block block.Block
	miner miner.Miner

	// 临时状态
	prevDiamondHash []byte

	// 上层 DB
	base *ChainState
}

var (
	globalInstanceChainStateMutex sync.Mutex
	globalInstanceChainState      *ChainState = nil
)

func GetGlobalInstanceChainState() *ChainState {
	globalInstanceChainStateMutex.Lock()
	defer globalInstanceChainStateMutex.Unlock()
	if globalInstanceChainState == nil {
		globalInstanceChainState = &ChainState{
			tempdir:   "",
			balanceDB: db.GetGlobalInstanceBalanceDB(),
			diamondDB: db.GetGlobalInstanceDiamondDB(),
		}
	}
	return globalInstanceChainState
}

func NewTempChainState(base *ChainState) *ChainState {
	if base == nil {
		base = GetGlobalInstanceChainState()
	}
	tmpdir := path.Join(os.TempDir(), "/hacash_state_temp_"+strconv.Itoa(rand.Int()))

	newBalanceDB := db.NewBalanceDB(path.Join(tmpdir, "balance"))
	newBalanceDB.Treedb.DeleteMark = true
	newBalanceDB.Treedb.FilePartitionLevel = 0 // 单文件

	newDiamondDB := db.NewDiamondDB(path.Join(tmpdir, "diamond"))
	newDiamondDB.Treedb.DeleteMark = true
	newDiamondDB.Treedb.FilePartitionLevel = 0 // 单文件

	// ok
	return &ChainState{
		base:      base,
		balanceDB: newBalanceDB,
		diamondDB: newDiamondDB,
		tempdir:   tmpdir,
		// 状态
		block:           nil,
		miner:           nil,
		prevDiamondHash: nil,
	}
}

// 区块
func (this *ChainState) Block() interface{} {
	// 正在处理的区块
	if this.block != nil {
		return this.block
	}
	// 查询上层
	if this.base != nil {
		return this.base.Block()
	}
	// 没有
	return nil
}

func (this *ChainState) SetBlock(blk interface{}) {
	this.block = blk.(block.Block)
}

func (this *ChainState) Miner() miner.Miner {
	return this.miner
}

func (this *ChainState) SetMiner(mir miner.Miner) {
	this.miner = mir
}

// 获取当前基于的钻石区块hash
func (this *ChainState) GetPrevDiamondHash() []byte {
	return this.prevDiamondHash
}

// 设置钻石区块hash
func (this *ChainState) SetPrevDiamondHash(hash []byte) {
	this.prevDiamondHash = hash
}

////////////////////////////////////////////////////////////

// copy
func (this *ChainState) TraversalCopy(get *ChainState) {
	this.balanceDB.Treedb.TraversalCopy(get.balanceDB.Treedb)
	this.diamondDB.Treedb.TraversalCopy(get.diamondDB.Treedb)
}

// 销毁临时状态
func (this *ChainState) Destroy() {
	//fmt.Println("os.RemoveAll " + this.tempdir)
	if this.tempdir != "" {
		os.RemoveAll(this.tempdir) // 删除所有文件
		this.tempdir = ""
	}
	//fmt.Println("os.RemoveAll --------- " + this.tempdir)
}
