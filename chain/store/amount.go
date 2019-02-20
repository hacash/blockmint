package store

import (
	"bytes"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
	"path"
)

var (
	BalanceDBItemMaxSize = uint32(5 + 32 + 23)
)

//
type StoreItemData struct {
	LockHeight fields.VarInt5 // 锁定区块高度
	BlankEmpty fields.Bytes32
	Amount     fields.Amount // 可用余额
	// cache data
	locitem *hashtreedb.IndexItem
}

func (this *StoreItemData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = this.LockHeight.Parse(buf, seek)
	seek, _ = this.BlankEmpty.Parse(buf, seek)
	seek, _ = this.Amount.Parse(buf, seek)
	return seek, nil
}

func (this *StoreItemData) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := this.LockHeight.Serialize()
	buffer.Write(b1)
	b2, _ := this.BlankEmpty.Serialize()
	buffer.Write(b2)
	b3, _ := this.Amount.Serialize()
	buffer.Write(b3)
	return buffer.Bytes(), nil
}

var (
	globalInstanceChainStateAmountDB *ChainStateBalanceDB = nil
)

//////////////////////////////////////////

type ChainStateBalanceDB struct {
	dirpath string

	treedb *hashtreedb.HashTreeDB

	delaycloseDBfile bool // 是否延迟关闭数据库

}

// 余额数据库全局实例
func GetGlobalInstanceChainStateBalanceDB() *ChainStateBalanceDB {
	if globalInstanceChainStateAmountDB != nil {
		return globalInstanceChainStateAmountDB
	}
	dir := config.GetCnfPathChainState()
	var db = new(ChainStateBalanceDB)
	db.Init(path.Join(dir, "balance"))
	globalInstanceChainStateAmountDB = db
	return db
}

func (this *ChainStateBalanceDB) Init(dir string) {
	this.delaycloseDBfile = false //
	this.dirpath = dir
	this.treedb = hashtreedb.NewHashTreeDB(dir, BalanceDBItemMaxSize, 20)
	this.treedb.FileName = "amt"
	this.treedb.FilePartitionLevel = 2
}

// 清空 重新创建
func (this *ChainStateBalanceDB) SaveAmountByClearCreate(address fields.Address, value *fields.Amount) error {
	storevalue := value.EllipsisDecimalFor23SizeStore() // ellipsis decimal
	var store StoreItemData
	store.Parse(fields.EmptyZeroBytes512, 0)
	store.Amount = *storevalue
	// save
	return this.Save(address, &store)
}

// 删除
func (this *ChainStateBalanceDB) Remove(address fields.Address) error {
	query, e1 := this.treedb.CreateQuery(address[1:]) // drop addr type
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
func (this *ChainStateBalanceDB) Delete(address fields.Address, store *StoreItemData) error {

	query, e1 := this.treedb.CreateQuery(address[1:]) // drop addr type
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// save
	e2 := query.Delete(store.locitem)
	if e2 != nil {
		return e2
	}
	// ok
	return nil
}

// 储存
func (this *ChainStateBalanceDB) Save(address fields.Address, store *StoreItemData) error {

	//fmt.Println( address[:] )
	query, e1 := this.treedb.CreateQuery(address[1:]) // drop addr type
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// save
	body, e2 := store.Serialize()
	if e2 != nil {
		return e2
	}
	var e3 error
	if store.locitem != nil {
		_, e3 = query.Write(store.locitem, body)
	} else {
		_, e3 = query.Save(body)
	}
	if e3 != nil {
		return e3
	}
	// ok
	return nil
}

// 读取
func (this *ChainStateBalanceDB) Read(address fields.Address) (*StoreItemData, error) {

	query, e1 := this.treedb.CreateQuery(address[1:]) // drop addr type
	if e1 != nil {
		return nil, e1
	}
	defer query.Close()
	// read
	result, item, e2 := query.Read()
	if e2 != nil {
		return nil, e2
	}
	if result == nil {
		return nil, nil // not find
	}
	if uint32(len(result)) < BalanceDBItemMaxSize {
		return nil, err.New("file size error")
	}
	var store StoreItemData
	_, e3 := store.Parse(result, 0)
	if e3 != nil {
		return nil, e3
	}
	store.locitem = item

	return &store, nil
}
