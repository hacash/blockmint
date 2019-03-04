package db

import (
	"bytes"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
	"path"
	"sync"
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
	Locitem *hashtreedb.IndexItem
}

func NewEmptyStoreItemData() *StoreItemData {
	return &StoreItemData{
		LockHeight: 0,
		BlankEmpty: fields.EmptyZeroBytes32,
		Amount:     *fields.NewEmptyAmount(),
	}
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

//////////////////////////////////////////

type BalanceDB struct {
	dirpath string

	Treedb *hashtreedb.HashTreeDB

	delaycloseDBfile bool // 是否延迟关闭数据库

}

var (
	globalInstanceBalanceDBMutex sync.Mutex
	globalInstanceBalanceDB      *BalanceDB = nil
)

// 余额数据库全局实例
func GetGlobalInstanceBalanceDB() *BalanceDB {
	globalInstanceBalanceDBMutex.Lock()
	defer globalInstanceBalanceDBMutex.Unlock()
	if globalInstanceBalanceDB != nil {
		return globalInstanceBalanceDB
	}
	dir := config.GetCnfPathChainState()
	var db = new(BalanceDB)
	db.Init(path.Join(dir, "balance"))
	globalInstanceBalanceDB = db
	return db
}

func NewBalanceDB(dir string) *BalanceDB {
	var db = new(BalanceDB)
	db.Init(dir)
	return db
}

func (this *BalanceDB) Init(dir string) {
	this.delaycloseDBfile = false //
	this.dirpath = dir
	this.Treedb = hashtreedb.NewHashTreeDB(dir, BalanceDBItemMaxSize, 20)
	this.Treedb.FileName = "amt"
	this.Treedb.FilePartitionLevel = 2
}

// 清空 重新创建
func (this *BalanceDB) SaveAmountByClearCreate(address fields.Address, value fields.Amount) error {
	storevalue := value.EllipsisDecimalFor23SizeStore() // ellipsis decimal
	var store = NewEmptyStoreItemData()
	store.Amount = *storevalue
	// save
	return this.Save(address, store)
}

// 删除
func (this *BalanceDB) Remove(address fields.Address) error {
	query, e1 := this.Treedb.CreateQuery(address[1:]) // drop addr type
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
func (this *BalanceDB) Delete(address fields.Address, store *StoreItemData) error {

	query, e1 := this.Treedb.CreateQuery(address[1:]) // drop addr type
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// save
	e2 := query.Delete(store.Locitem)
	if e2 != nil {
		return e2
	}
	// ok
	return nil
}

// 储存
func (this *BalanceDB) Save(address fields.Address, store *StoreItemData) error {

	//fmt.Println( address[:] )
	query, e1 := this.Treedb.CreateQuery(address[1:]) // drop addr type
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
	if store.Locitem != nil {
		_, e3 = query.Write(store.Locitem, body)
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
func (this *BalanceDB) Read(address fields.Address) (*StoreItemData, error) {

	query, e1 := this.Treedb.CreateQuery(address[1:]) // drop addr type
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
	store.Locitem = item

	return &store, nil
}
