package store

import (
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
)

var (
	trsIdxItemSizeSet = uint32(1 + 4 + 3*4)
)

//////////////////////////////////////////////////////

type TrsIdxDB struct {
	filepath string

	treedb *hashtreedb.HashTreeDB
}

func (this *TrsIdxDB) Init(filepath string) {
	this.filepath = filepath
	this.treedb = hashtreedb.NewHashTreeDB(filepath, trsIdxItemSizeSet, 32)
	this.treedb.FilePartitionLevel = 4 // 文件分区层级
	this.treedb.FileName = "TRS"
}

func (this *TrsIdxDB) Save(hash []byte, saveval *TrsIdxOneFindItem) (*hashtreedb.IndexItem, error) {
	query, e := this.treedb.CreateQuery(hash)
	if e != nil {
		return nil, e
	}
	// save
	item, e1 := query.Save(saveval.Serialize())
	if e1 != nil {
		return nil, e1
	}
	query.Close()
	// ok
	return item, nil
}

func (this *TrsIdxDB) Find(hash []byte) (*TrsIdxOneFindItem, error) {
	query, e1 := this.treedb.CreateQuery(hash)
	if e1 != nil {
		return nil, e1
	}
	result, e2 := query.Read()
	if e2 != nil {
		return nil, e2
	}
	if uint32(len(result)) < trsIdxItemSizeSet {
		return nil, err.New("file store error")
	}

	var item TrsIdxOneFindItem
	item.Parse(result, 0)
	if item.location == nil {
		return nil, nil
	}

	return &item, nil
}
