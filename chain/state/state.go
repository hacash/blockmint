package state

import (
	"github.com/hacash/blockmint/chain/state/db"
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
	//fmt.Println("os.RemoveAll " + this.tempdir)
	if this.tempdir != "" {
		os.RemoveAll(this.tempdir) // 删除所有文件
		this.tempdir = ""
	}
	//fmt.Println("os.RemoveAll --------- " + this.tempdir)
}
