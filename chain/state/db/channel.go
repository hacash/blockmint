package db

import (
	"bytes"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/service/hashtreedb"
	"path"
	"sync"
)

/**
 * 支付通道储存
 */

var (
	ChannelDBItemMaxSize = uint32(5 + 2 + (21+6)*2 + 3 + 16) // 80 = 16 × 5
)

//
type ChannelStoreItemData struct {
	BelongHeight fields.VarInt5 // 通道开启时的区块高度
	LockBlock    fields.VarInt2 // 单方面结束通道要锁定的区块数量
	LeftAddress  fields.Address
	LeftAmount   fields.Amount // 抵押数额1
	RightAddress fields.Address
	RightAmount  fields.Amount  // 抵押数额2
	IsClosed     fields.VarInt1 // 已经关闭并结算
	ConfigMark   fields.VarInt2 // 标志位
	Others       fields.Bytes16 // 扩展位

	// cache data
	Locitem *hashtreedb.IndexItem
}

func (this *ChannelStoreItemData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = this.BelongHeight.Parse(buf, seek)
	seek, _ = this.LockBlock.Parse(buf, seek)
	seek, _ = this.LeftAddress.Parse(buf, seek)
	seek, _ = this.LeftAmount.Parse(buf, seek)
	seek, _ = this.RightAddress.Parse(buf, seek)
	seek, _ = this.RightAmount.Parse(buf, seek)
	seek, _ = this.IsClosed.Parse(buf, seek)
	seek, _ = this.ConfigMark.Parse(buf, seek)
	seek, _ = this.Others.Parse(buf, seek)
	return seek, nil
}

func (this *ChannelStoreItemData) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := this.LockBlock.Serialize()
	b2, _ := this.BelongHeight.Serialize()
	b3, _ := this.LeftAddress.Serialize()
	b4, _ := this.LeftAmount.Serialize()
	b5, _ := this.RightAddress.Serialize()
	b6, _ := this.RightAmount.Serialize()
	b7, _ := this.IsClosed.Serialize()
	b8, _ := this.ConfigMark.Serialize()
	b9, _ := this.Others.Serialize()
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	buffer.Write(b4)
	buffer.Write(b5)
	buffer.Write(b6)
	buffer.Write(b7)
	buffer.Write(b8)
	buffer.Write(b9)
	return buffer.Bytes(), nil
}

//////////////////////////////////////////

var (
	globalInstanceChannelDBMutex sync.Mutex
	globalInstanceChannelDB      *ChannelDB = nil
)

// 余额数据库全局实例
func GetGlobalInstanceChannelDB() *ChannelDB {
	globalInstanceChannelDBMutex.Lock()
	defer globalInstanceChannelDBMutex.Unlock()
	if globalInstanceChannelDB != nil {
		return globalInstanceChannelDB
	}
	dir := config.GetCnfPathChainState()
	var db = NewChannelDB(path.Join(dir, "channel"))
	globalInstanceChannelDB = db
	return db
}

type ChannelDB struct {
	dirpath string

	Treedb *hashtreedb.HashTreeDB
}

func NewChannelDB(dir string) *ChannelDB {
	var db = new(ChannelDB)
	db.Init(dir)
	return db
}

func (this *ChannelDB) Init(dir string) {
	this.dirpath = dir
	this.Treedb = hashtreedb.NewHashTreeDB(dir, ChannelDBItemMaxSize, 16)
	this.Treedb.FileName = "chl"
	this.Treedb.FilePartitionLevel = 1
}

// 读取
func (this *ChannelDB) Read(channelNumber fields.Bytes16) (*ChannelStoreItemData, error) {

	query, e1 := this.Treedb.CreateQuery(channelNumber) // drop addr type
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
	var item ChannelStoreItemData
	if _, err := item.Parse(result, 0); err != nil {
		return nil, err
	}
	// 返回
	return &item, nil
}

// 删除
func (this *ChannelDB) Delete(channelNumber fields.Bytes16) error {

	query, e1 := this.Treedb.CreateQuery(channelNumber)
	if e1 != nil {
		return e1
	}
	defer query.Close()
	// delete
	e2 := query.Remove()
	if e2 != nil {
		return e2
	}
	// ok
	return nil
}

// 储存
func (this *ChannelDB) Save(channelNumber fields.Bytes16, store *ChannelStoreItemData) error {

	//fmt.Println( address[:] )
	query, e1 := this.Treedb.CreateQuery(channelNumber) // drop addr type
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
