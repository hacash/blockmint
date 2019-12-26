package store

import (
	"bytes"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/types/block"
)

var (
	valueSizeSet = BlockLocationSize + uint32(block1def.ByteSizeBlockHead)
)

type BlockIndexDB struct {
	filepath string
	treedb   *hashtreedb.HashTreeDB
}

func (this *BlockIndexDB) Init(filepath string) {
	this.filepath = filepath
	this.treedb = hashtreedb.NewHashTreeDB(filepath, valueSizeSet, 32)
	this.treedb.KeyReverse = true      // key值倒序
	this.treedb.FilePartitionLevel = 2 // 文件分区
}

func (this *BlockIndexDB) Save(hash []byte, blockLoc *BlockLocation, block block.Block) (*hashtreedb.IndexItem, error) {
	blockheadbytes, err := block.SerializeHead()
	if err != nil {
		return nil, err
	}
	return this.SaveByBlockHeadByte(hash, blockLoc, blockheadbytes)
}

func (this *BlockIndexDB) SaveByBlockHeadByte(hash []byte, blockLoc *BlockLocation, blockheadbytes []byte) (*hashtreedb.IndexItem, error) {

	query, e := this.treedb.CreateQuery(hash)
	defer query.Close()
	if e != nil {
		return nil, e
	}
	// new body
	var bodybyte bytes.Buffer
	bodybyte.Write(blockLoc.Serialize())
	bodybyte.Write(blockheadbytes)
	// save
	item, e2 := query.Save(bodybyte.Bytes())
	if e2 != nil {
		return nil, e2
	}
	// ok
	return item, nil
}

func (this *BlockIndexDB) Find(hash []byte) (*BlockLocation, block.Block, error) {
	loc, hdbytes, e := this.FindBlockHeadBytes(hash)
	if e != nil {
		return nil, nil, e
	}
	if hdbytes == nil {
		return nil, nil, nil // not find
	}
	var block, _, e3 = blocks.ParseBlockHead(hdbytes, 0)
	if e3 != nil {
		return nil, nil, e3
	}
	return loc, block, nil

}

func (this *BlockIndexDB) FindBlockHeadBytes(hash []byte) (*BlockLocation, []byte, error) {
	query, e1 := this.treedb.CreateQuery(hash)
	defer query.Close()
	if e1 != nil {
		return nil, nil, e1
	}
	// read
	result, _, e2 := query.Read()
	if e2 != nil {
		return nil, nil, e2
	}
	if result == nil {
		return nil, nil, nil // not find
	}
	if uint32(len(result)) < valueSizeSet {
		return nil, nil, err.New("file store error")
	}

	var loc BlockLocation
	loc.Parse(result, 0)
	//var block, _, e3 = blocks.ParseBlockHead(result, 3*4)
	//if e3 != nil {
	//	return nil, nil, e3
	//}
	start := 3 * 4
	return &loc, result[start : start+block1def.ByteSizeBlockHead], nil
}

func (this *BlockIndexDB) FindBlockHashByPosition(keyprefix []byte, ptrnum uint32) ([]byte, error) {
	hashsize := int64(32)
	valuebytes, e := this.treedb.ReadBytesByPositionWithLength(keyprefix, ptrnum, hashsize)
	if e != nil {
		return nil, e
	}
	return valuebytes, nil
}

func (this *BlockIndexDB) FindBlockHeadBytesByPosition(keyprefix []byte, ptrnum uint32) ([]byte, []byte, error) {
	valuebytes, e := this.treedb.ReadBytesByPosition(keyprefix, ptrnum)
	if e != nil {
		return nil, nil, e
	}
	start := this.treedb.HashSize + BlockLocationSize
	return valuebytes[0:this.treedb.HashSize], valuebytes[start : start+uint32(block1def.ByteSizeBlockHead)], nil
}

/////////////////////////////////////////

func (this *BlockIndexDB) GetPositionLvTwoByHash(hash []byte) [2]byte {
	hashkey := hash
	if this.treedb.KeyReverse {
		hashkey = hashtreedb.ReverseHashOrder(hash) // 倒序
	}
	return [2]byte{hashkey[0], hashkey[1]}

}
