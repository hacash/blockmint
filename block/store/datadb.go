package store

import (
	"github.com/hacash/blockmint/protocol/blockdef"
	"os"
	"path"
	"strconv"

	//"fmt"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	//"os"
)

var (
	maxPartFileSize = 1 // 400 MB

)

type BlockDataDB struct {
	filepath string

	fileHeadName  string
	fileIndexName string

	fileHead *os.File

	blkf BlockStoreFileHead

	indexdb BlockIndexDB
}

func (db *BlockDataDB) getPartFileName(filenum uint32) string {
	partnum := strconv.Itoa(int(filenum))
	return path.Join(db.filepath, "part_"+string(partnum)+".dat")
}

func (db *BlockDataDB) Init(filepath string) {

	db.filepath = filepath
	db.fileHeadName = path.Join(filepath, "HEAD.dat")
	db.fileIndexName = path.Join(filepath, "INDEX.dat")
	file.CreatePath(db.filepath)
	db.indexdb.Init(db.fileIndexName)
	//db.fileHead, _ = os.OpenFile(db.fileHeadName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	//db.fileIndex, _ = os.OpenFile(db.fileIndexName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空

	db.blkf.Load(db.fileHeadName) // load

	defer func() {
		//db.fileHead.Close()
		//db.fileIndex.Close()
		//db.fileHead = nil
		//db.fileIndex = nil
	}()

}

func (db *BlockDataDB) glowBlockStoreFileNum() error {
	db.blkf.FileNum += 1
	db.blkf.Flush(db.fileHeadName)
	return nil
}

func (db *BlockDataDB) SaveBlock(blockhash []byte, block block.Block) error {
	blkbytes, _ := block.Serialize()
	return db.SaveBlockByBytes(blockhash, blkbytes)
}

func (db *BlockDataDB) SaveBlockByBytes(blockhash []byte, blockbytes []byte) error {

	head := blockbytes[:blockdef.ByteSizeBlockHead]
	blockbody := blockbytes[blockdef.ByteSizeBlockHead:]
	bodyLen := len(blockbody)

	var filenum = db.blkf.FileNum
	var curFileName = db.getPartFileName(filenum)
	var nextFileName = db.getPartFileName(filenum + 1)

	currentBlockDataFile, _ := os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	filestat, e := currentBlockDataFile.Stat()
	if e != nil {
		return nil
	}

	var curFileSize = filestat.Size()
	if curFileSize+int64(bodyLen) > int64(maxPartFileSize)*1024*1024 {
		currentBlockDataFile, _ = os.OpenFile(nextFileName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
		db.glowBlockStoreFileNum()                                                       // num + 1
		curFileSize = 0
	}

	var location = &BlockLocation{
		db.blkf.FileNum,
		uint32(curFileSize),
		uint32(bodyLen),
	}

	// do store data
	currentBlockDataFile.WriteAt(blockbody, curFileSize)

	// do store index
	db.indexdb.SaveByByteForce(blockhash, location, head)

	return nil

}

func (db *BlockDataDB) ReadBlock(blockhash []byte) block.Block {

	// read index
	item := db.indexdb.Find(blockhash)
	if item == nil {
		return nil
	}
	loc := item.Location

	// read body
	tarFileName := db.getPartFileName(loc.BlockFileNum)
	tarFile, _ := os.OpenFile(tarFileName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空

	var bodyBytes = make([]byte, loc.BlockLen)
	rdlen, e := tarFile.ReadAt(bodyBytes, int64(loc.FileOffset))
	if e != nil || uint32(rdlen) != loc.BlockLen {
		return nil
	}

	// parse
	item.BlockHead.ParseBody(bodyBytes, 0)

	return item.BlockHead

}
