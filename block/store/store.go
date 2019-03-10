package store

import (
	"bytes"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	"path"
	"sync"
)

type BlocksDataStore struct {
	basedir string

	datadb  BlockDataDB
	indexdb BlockIndexDB
	heidxdb BlockHeightDB
	trsdb   TrsIdxDB
}

var (
	blockStoreInstanceMutex sync.Mutex
	blockStoreInstance      *BlocksDataStore = nil
)

func GetGlobalInstanceBlocksDataStore() *BlocksDataStore {
	blockStoreInstanceMutex.Lock()
	defer blockStoreInstanceMutex.Unlock()
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
		/*
			if this.total > 1 {
				fmt.Println("Save trs "+hex.EncodeToString(trs.HashNoFee()))
			}
			if this.total > 1 {
				fmt.Println(trs.HashNoFee())
			}
		*/
		_, e := this.db.trsdb.Save(trs.HashNoFee(), &TrsIdxOneFindItem{
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
	this.heidxdb.Init(path.Join(dir, "heidxs"))
	this.trsdb.Init(path.Join(dir, "trsidxs"))
	this.datadb.Init(dir)

	return nil
}

// 储存 合法的 区块
func (this *BlocksDataStore) Save(blk block.Block) ([]byte, error) {
	// hash
	blockhash := blk.Hash()
	// trsidx
	itr := &FinishOneTrsItIteratorSaveTrs{
		db: this,
	}
	trsbytes, e3 := blk.SerializeTransactions(itr)
	if e3 != nil {
		return nil, e3
	}
	// head
	blockheadbytes, e1 := blk.SerializeHead()
	blockmetabytes, e2 := blk.SerializeMeta()
	if e2 != nil {
		return nil, e2
	}
	// data
	var blockbodybuf bytes.Buffer
	blockbodybuf.Write(blockmetabytes)
	blockbodybuf.Write(trsbytes)
	blockbodys := blockbodybuf.Bytes()
	// save body
	blksaveloc, e1 := this.datadb.SaveByBodyBytes(blockbodys)
	if e1 != nil {
		return nil, e1
	}
	// index db
	item, e5 := this.indexdb.SaveByBlockHeadByte(blockhash, blksaveloc, blockheadbytes)
	if e5 != nil {
		return nil, e5
	}
	blkhdfilepart := this.indexdb.GetPositionLvTwoByHash(blockhash)
	blkhdptrnum := item.ValuePtrNum
	// save height index
	e6 := this.heidxdb.Save(blk.GetHeight(), blkhdfilepart, blkhdptrnum, blksaveloc)
	if e6 != nil {
		return nil, e6
	}
	// save trs
	e7 := itr.SaveAll(blkhdfilepart, blkhdptrnum, blksaveloc)
	if e7 != nil {
		return nil, e7
	}
	var blkallbytes bytes.Buffer
	blkallbytes.Write(blockheadbytes)
	blkallbytes.Write(blockbodys)
	return blkallbytes.Bytes(), nil
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
	var resbuf bytes.Buffer
	resbuf.Write(blkhead)
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

// 检查交易是否存在
func (this *BlocksDataStore) CheckTransactionExist(hashNoFee []byte) (bool, error) {
	res, e := this.ReadTransaction(hashNoFee, false, false)
	if e != nil {
		return false, e
	}
	if res != nil && res.Location != nil && res.Location.DataLen > uint32(0) {
		return true, nil
	}
	return false, nil

}

func (this *BlocksDataStore) GetBlockHashByHeight(height uint64) ([]byte, error) {
	finditem, e := this.heidxdb.Find(height)
	if e != nil {
		return nil, e
	}
	if finditem == nil {
		return nil, nil
	}

	hash, e1 := this.indexdb.FindBlockHashByPosition(finditem.BlockHeadInfoFilePartition[:], finditem.BlockHeadInfoPtrNumber)
	if e1 != nil {
		return nil, e1
	}

	return hash, nil
}

func (this *BlocksDataStore) GetBlockBytesByHeight(height uint64, gethead bool, getbody bool) ([]byte, error) {
	finditem, e := this.heidxdb.Find(height)
	if e != nil {
		return nil, e
	}
	if finditem == nil {
		return nil, nil
	}
	var blockbytes bytes.Buffer
	if gethead {
		head, e1 := this.indexdb.FindBlockHeadBytesByPosition(finditem.BlockHeadInfoFilePartition[:], finditem.BlockHeadInfoPtrNumber)
		//fmt.Println(e1)
		if e1 != nil {
			return nil, e1
		}
		blockbytes.Write(head)
	}
	if getbody {
		body, e := this.datadb.ReadBlockBody(finditem.location)
		//fmt.Println(body)
		if e != nil {
			return nil, e
		}
		blockbytes.Write(body)
	}
	// ok
	return blockbytes.Bytes(), nil
}

// 强制非安全删除区块、交易等数据
func (this *BlocksDataStore) DeleteBlockForceUnsafe(block block.Block) error {
	// 循环删除交易指针即可
	for _, tx := range block.GetTransactions() {
		this.trsdb.Delete(tx.HashNoFee())
	}
	return nil
}
