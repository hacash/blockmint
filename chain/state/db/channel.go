package db

import "github.com/hacash/blockmint/service/hashtreedb"

/**
 * 支付通道储存
 */

//////////////////////////////////////////

var (
	ChannelDBItemMaxSize = uint32(5 + 4 + (21+6)*2)
)

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
