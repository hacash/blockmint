package store

import (
	"bytes"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	"path"
)

type BlocksDataStore struct {
	basedir string

	datadb  BlockDataDB
	indexdb BlockIndexDB
	trsdb   TrsIdxDB
}

var blockStoreInstance *BlocksDataStore = nil

func GetBlocksDataStoreInstance() *BlocksDataStore {
	if blockStoreInstance != nil {
		return blockStoreInstance
	}
	blockStoreInstance = new(BlocksDataStore)
	blockStoreInstance.Init(config.GetCnfPathBlocks())
	return blockStoreInstance
}

type FinishOneTrsItIteratorSaveTrs struct {
	db *BlocksDataStore

	total     uint32
	trslist   []block.Transaction
	byteslist [][]byte
}

func (this *FinishOneTrsItIteratorSaveTrs) Init(total uint32) {
	this.total = total
	this.trslist = make([]block.Transaction, total)
	this.byteslist = make([][]byte, total)
}
func (this *FinishOneTrsItIteratorSaveTrs) FinishOneTrs(i uint32, trs block.Transaction, trsbytes []byte) {
	this.trslist[i] = trs
	this.byteslist[i] = trsbytes
}
func (this *FinishOneTrsItIteratorSaveTrs) SaveAll(partnum [2]byte, headptrnum uint32, blkloc *BlockLocation) error {
	appendOffset := uint32(block1def.ByteSizeBlockMeta)
	for i := uint32(0); i < this.total; i++ {
		trs := this.trslist[i]
		trslen := uint32(len(this.byteslist[i]))
		_, e := this.db.trsdb.Save(trs.Hash(), &TrsIdxOneFindItem{
			partnum,
			headptrnum,
			&BlockLocation{
				blkloc.BlockFileNum,
				blkloc.FileOffset + appendOffset,
				trslen,
			},
		})
		if e != nil {
			return e
		}
		appendOffset += trslen
	}
	return nil
}

//////////////

func (this *BlocksDataStore) Init(dir string) error {
	this.basedir = dir
	file.CreatePath(dir)

	this.indexdb.Init(path.Join(dir, "indexs"))
	this.trsdb.Init(path.Join(dir, "trsidxs"))
	this.datadb.Init(dir)

	return nil
}

// 储存 合法的 区块
func (this *BlocksDataStore) Save(blk block.Block) error {
	// hash
	blockhash := blk.Hash()
	// trsidx
	itr := &FinishOneTrsItIteratorSaveTrs{
		db: this,
	}
	trsbytes, e3 := blk.SerializeTransactions(itr)
	if e3 != nil {
		return e3
	}
	// head
	blockheadbytes, e1 := blk.SerializeHead()
	blockmetabytes, e2 := blk.SerializeMeta()
	if e2 != nil {
		return e2
	}
	// data
	blockbodybuf := bytes.NewBuffer(blockmetabytes)
	blockbodybuf.Write(trsbytes)
	// save body
	loc, e1 := this.datadb.SaveByBodyBytes(blockbodybuf.Bytes())
	if e1 != nil {
		return e1
	}
	item, e5 := this.indexdb.SaveByBlockHeadByte(blockhash, loc, blockheadbytes)
	if e5 != nil {
		return e5
	}
	// save trs
	e4 := itr.SaveAll(this.indexdb.GetPositionLvTwoByHash(blockhash), item.ValuePtrNum, loc)
	if e4 != nil {
		return e4
	}

	return nil
}

func (this *BlocksDataStore) Read(hash []byte) (block.Block, error) {
	blockbytes, e := this.ReadBlockBytes(hash)
	if e != nil {
		return nil, e
	}
	blk, _, e1 := blocks.ParseBlock(blockbytes, 0)
	if e1 != nil {
		return nil, e1
	}
	return blk, nil
}

func (this *BlocksDataStore) ReadBlockBytes(hash []byte) ([]byte, error) {
	blkloc, blkhead, e := this.indexdb.FindBlockHeadBytes(hash)
	if e != nil {
		return nil, e
	}
	var resbuf = bytes.NewBuffer(blkhead)
	blkbodybytes, e1 := this.datadb.ReadBlockBody(blkloc)
	if e1 != nil {
		return nil, e1
	}
	resbuf.Write(blkbodybytes)
	return resbuf.Bytes(), nil
}

func (this *BlocksDataStore) ReadHead(hash []byte) (block.Block, error) {
	_, blkhead, e := this.indexdb.Find(hash)
	if e != nil {
		return nil, e
	}
	return blkhead, nil
}

func (this *BlocksDataStore) ReadTransaction(hash []byte, getbody bool, getblockhead bool) (*TrsFindResult, error) {

	finditem, e := this.trsdb.Find(hash)
	if e != nil {
		return nil, e
	}
	if finditem == nil {
		return nil, nil
	}
	var restrs TrsFindResult
	restrs.Location = finditem.location

	if getbody {
		body, e := this.datadb.ReadBlockBody(finditem.location)
		//fmt.Println(body)
		if e != nil {
			return nil, e
		}
		trs, _, e1 := blocks.ParseTransaction(body, 0)
		if e1 != nil {
			return nil, e1
		}
		restrs.Transaction = trs
	}
	if getblockhead {
		head, e1 := this.indexdb.FindBlockHeadBytesByPosition(finditem.BlockHeadInfoFilePartition[:], finditem.BlockHeadInfoPtrNumber)
		//fmt.Println(head)
		if e1 != nil {
			return nil, e1
		}
		blkhd, _, e2 := blocks.ParseBlockHead(head, 0)
		if e2 != nil {
			return nil, e2
		}
		restrs.BlockHead = blkhd
	}
	// ok
	return &restrs, nil
}
