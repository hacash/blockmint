package db

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/service/hashtreedb"
	"path"
	"sync"
)

/**
 * 支付通道储存
 */

//////////////////////////////////////////

var (
	DiamondDBItemMaxSize = uint32(21)
)

type DiamondDB struct {
	dirpath string

	Treedb *hashtreedb.HashTreeDB
}

var (
	globalInstanceDiamondDBMutex sync.Mutex
	globalInstanceDiamondDB      *DiamondDB = nil
)

// 余额数据库全局实例
func GetGlobalInstanceDiamondDB() *DiamondDB {
	globalInstanceDiamondDBMutex.Lock()
	defer globalInstanceDiamondDBMutex.Unlock()
	if globalInstanceDiamondDB != nil {
		return globalInstanceDiamondDB
	}
	dir := config.GetCnfPathChainState()
	var db = new(DiamondDB)
	db.Init(path.Join(dir, "diamond"))
	globalInstanceDiamondDB = db
	return db
}

func NewDiamondDB(dir string) *DiamondDB {
	var db = new(DiamondDB)
	db.Init(dir)
	return db
}

func (this *DiamondDB) Init(dir string) {
	this.dirpath = dir
	this.Treedb = hashtreedb.NewHashTreeDB(dir, DiamondDBItemMaxSize, 6)
	this.Treedb.FileName = "dmd"
	this.Treedb.FilePartitionLevel = 1
}

// 读取
func (this *DiamondDB) Read(diamond fields.Bytes6) (fields.Address, error) {

	query, e1 := this.Treedb.CreateQuery(diamond) // drop addr type
	if e1 != nil {
		return nil, e1
	}
	defer query.Close()
	// read
	result, _, e2 := query.Read()
	if e2 != nil {
		return nil, e2
	}
	if result == nil {
		return nil, nil // not find
	}
	// 返回所属地址
	return result, nil
}

// 储存
func (this *DiamondDB) SetBelong(diamond fields.Bytes6, address fields.Address) error {
	//fmt.Println( address[:] )
	query, e1 := this.Treedb.CreateQuery(diamond) // drop addr type
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// save
	_, e3 := query.Save(address)
	if e3 != nil {
		return e3
	}
	// ok
	return nil
}

// 删除
func (this *DiamondDB) Delete(diamond fields.Bytes6) error {
	query, e1 := this.Treedb.CreateQuery(diamond) // drop addr type
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// save
	e2 := query.Remove()
	if e2 != nil {
		return e2
	}
	// ok
	return nil
}
