package store

import (
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/err"
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

	fileHeadName string

	fileHead *os.File

	blkf BlockStoreFileHead
}

func (db *BlockDataDB) getPartFileName(filenum uint32) string {
	partnum := strconv.Itoa(int(filenum))
	return path.Join(db.filepath, "blk"+string(partnum)+".dat")
}

func (db *BlockDataDB) Init(filepath string) {

	db.filepath = filepath
	db.fileHeadName = path.Join(filepath, "HEAD.hd")
	file.CreatePath(db.filepath)

	db.blkf.Load(db.fileHeadName) // load

}

func (db *BlockDataDB) glowBlockStoreFileNum() error {
	db.blkf.FileNum += 1
	db.blkf.Flush(db.fileHeadName)
	return nil
}

func (db *BlockDataDB) Save(block block.Block) (*BlockLocation, error) {
	blkbytes, _ := block.Serialize()
	blockbody := blkbytes[block1def.ByteSizeBlockHead:]
	return db.SaveByBodyBytes(blockbody)
}

func (db *BlockDataDB) SaveByBodyBytes(blockbody []byte) (*BlockLocation, error) {

	bodyLen := len(blockbody)

	var filenum = db.blkf.FileNum
	var curFileName = db.getPartFileName(filenum)
	var nextFileName = db.getPartFileName(filenum + 1)

	currentBlockDataFile, _ := os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	filestat, e := currentBlockDataFile.Stat()
	if e != nil {
		return nil, e
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
	_, er := currentBlockDataFile.WriteAt(blockbody, curFileSize)
	if er != nil {
		return nil, er
	}

	return location, nil

}

func (db *BlockDataDB) ReadBlockBody(loc *BlockLocation) ([]byte, error) {

	// read body
	tarFileName := db.getPartFileName(loc.BlockFileNum)
	tarFile, _ := os.OpenFile(tarFileName, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空

	var bodyBytes = make([]byte, loc.DataLen)
	rdlen, e := tarFile.ReadAt(bodyBytes, int64(loc.FileOffset))
	if e != nil {
		return nil, e
	}
	if uint32(rdlen) != loc.DataLen {
		return nil, err.New("error file size")
	}

	return bodyBytes, nil
}
