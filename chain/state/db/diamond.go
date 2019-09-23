package db

import (
	"bytes"
	"crypto/md5"
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
	DiamondDBItemMaxSize = uint32(5 + 3 + 21)
)

type DiamondDB struct {
	dirpath string

	Treedb *hashtreedb.HashTreeDB
}

var (
	globalInstanceDiamondDBMutex sync.Mutex
	globalInstanceDiamondDB      *DiamondDB = nil
)

//////////////////////////////////////////

type DiamondStoreItemData struct { // len = 5 + 3 + 21 = 29
	BlockHeight fields.VarInt5
	Number      fields.VarInt3
	Address     fields.Address
}

func (this *DiamondStoreItemData) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = this.BlockHeight.Parse(buf, seek)
	seek, _ = this.Number.Parse(buf, seek)
	seek, _ = this.Address.Parse(buf, seek)
	return seek, nil
}

func (this *DiamondStoreItemData) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	b1, _ := this.BlockHeight.Serialize()
	b2, _ := this.Number.Serialize()
	b3, _ := this.Address.Serialize()
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	return buffer.Bytes(), nil
}

//////////////////////////////////////////

func dealDiamondQueryKey(diamond fields.Bytes6) []byte {
	md5hx := md5.Sum(diamond)
	//fmt.Println(diamond, md5hx)
	buf := bytes.NewBuffer(md5hx[0:4])
	buf.Write(diamond)
	return buf.Bytes()
}

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
	this.Treedb = hashtreedb.NewHashTreeDB(dir, DiamondDBItemMaxSize, 10) // 4 + 6
	this.Treedb.FileName = "dmd"
	this.Treedb.FilePartitionLevel = 2
}

// 读取
func (this *DiamondDB) Read(diamond fields.Bytes6) (*DiamondStoreItemData, error) {
	dmdkey := dealDiamondQueryKey(diamond)
	//fmt.Println(dmdkey)
	query, e1 := this.Treedb.CreateQuery(dmdkey)
	defer query.Close()
	if e1 != nil {
		return nil, e1
	}
	// read
	result, _, e2 := query.Read()
	if e2 != nil {
		return nil, e2
	}
	if result == nil {
		//fmt.Println("	result, _, e2 := query.Read()   notfind 2   " + string(diamond) +"   "+ string(dmdkey))
		return nil, nil // not find
	}
	// 返回
	var item = &DiamondStoreItemData{}
	item.Parse(result, 0)
	return item, nil
}

// 创建 或 修改所属
func (this *DiamondDB) Save(diamond fields.Bytes6, storeitem *DiamondStoreItemData) error {
	//fmt.Println( address[:] )
	dmdkey := dealDiamondQueryKey(diamond)
	query, e1 := this.Treedb.CreateQuery(dmdkey)
	defer query.Close()
	if e1 != nil {
		return e1
	}
	// save
	diamonditembuf, _ := storeitem.Serialize()
	_, e3 := query.Save(diamonditembuf)
	if e3 != nil {
		return e3
	}
	// 测试查询
	//query.Close()
	//ddd, eee := this.Read( diamond )
	////fmt.Println(ddd, eee, ddd.Address.ToReadable())
	//
	//if strings.Compare(string(diamond), "NHMYYM") == 0{
	//	//panic("NHMYYM")
	//}

	// ok
	return nil
}

// 删除
func (this *DiamondDB) Delete(diamond fields.Bytes6) error {
	dmdkey := dealDiamondQueryKey(diamond)
	query, e1 := this.Treedb.CreateQuery(dmdkey)
	defer query.Close()
	if e1 != nil {
		return e1
	}
	// save
	e2 := query.Remove()
	if e2 != nil {
		return e2
	}
	// ok
	return nil
}
