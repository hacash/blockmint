package state

import (
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/config"
	"math/rand"
	"os"
	"path"
	"strconv"
)

//
type ChainState struct {
	tempdir string

	balanceDB *db.BalanceDB

	// 上层 DB
	base *ChainState
}

var (
	globalInstanceChainState *ChainState = nil
)

func GetGlobalInstanceChainState() *ChainState {
	if globalInstanceChainState == nil {
		globalInstanceChainState = &ChainState{
			tempdir:   "",
			balanceDB: db.GetGlobalInstanceBalanceDB(),
		}
	}
	return globalInstanceChainState
}

func NewTempChainState(base *ChainState) *ChainState {
	if base == nil {
		base = GetGlobalInstanceChainState()
	}
	tmpdir := config.GetCnfPathMinerState() + "/temp" + strconv.Itoa(rand.Int())

	newBalanceDB := db.NewBalanceDB(path.Join(tmpdir, "balance"))
	newBalanceDB.Treedb.DeleteMark = true
	newBalanceDB.Treedb.FilePartitionLevel = 0 // 单文件

	// ok
	return &ChainState{
		base:      base,
		balanceDB: newBalanceDB,
		tempdir:   tmpdir,
	}
}

// copy
func (this *ChainState) TraversalCopy(get *ChainState) {
	this.balanceDB.Treedb.TraversalCopy(get.balanceDB.Treedb)
}

///////////////////////

// 销毁临时状态
func (this *ChainState) Destroy() {
	if this.tempdir != "" {
		os.RemoveAll(this.tempdir) // 删除所有文件
	}
}
