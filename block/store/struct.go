package store

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/types/block"
	"os"
)

type BlockStoreFileHead struct {
	FileNum uint32
}

func (loc *BlockStoreFileHead) wideSize() uint32 {
	return 4
}

func (this *BlockStoreFileHead) Load(filepath string) error {
	fileHead, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	var size = this.wideSize()
	var buffer = make([]byte, size)
	rdlen, e := fileHead.ReadAt(buffer, 0)
	if e != nil || rdlen != int(size) {
		return err.New("file read error")
	}
	this.Parse(buffer, 0)
	defer func() {
		fileHead.Close()
	}()
	return nil
}

func (this *BlockStoreFileHead) Flush(filepath string) error {
	fileHead, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空

	var buffer = this.Serialize()

	fileHead.WriteAt(buffer, 0)

	defer func() {
		fileHead.Close()
	}()
	return nil
}

func (loc *BlockStoreFileHead) Parse(buf []byte, seek uint32) error {
	loc.FileNum = binary.BigEndian.Uint32(buf[seek : seek+4])
	return nil
}

func (loc *BlockStoreFileHead) Serialize() []byte {
	var byt1 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt1, loc.FileNum)
	var buffer bytes.Buffer
	buffer.Write(byt1)
	return buffer.Bytes()
}

////////////////////////////////////////////////////

type BlockLocation struct {
	BlockFileNum uint32
	FileOffset   uint32
	DataLen      uint32
}

func (loc BlockLocation) Size() uint32 {
	return 3 * 4
}

func (loc *BlockLocation) Parse(buf []byte, seek uint32) error {
	loc.BlockFileNum = binary.BigEndian.Uint32(buf[seek : seek+4])
	loc.FileOffset = binary.BigEndian.Uint32(buf[seek+4 : seek+8])
	loc.DataLen = binary.BigEndian.Uint32(buf[seek+8 : seek+12])
	return nil
}

func (loc *BlockLocation) Serialize() []byte {
	var byt1 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt1, loc.BlockFileNum)
	var byt2 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt2, loc.FileOffset)
	var byt3 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt3, loc.DataLen)
	var buffer bytes.Buffer
	buffer.Write(byt1)
	buffer.Write(byt2)
	buffer.Write(byt3)
	return buffer.Bytes()
}

////////////////////////////////////////////////////

type PositionFindItem struct {
	itemoffset int64  // 指针储存位置
	itemkind   uint8  // 数据种类 0 1 2
	itemnumber uint32 // 数据位置编号
	// body
	hashtail  []byte
	Blockhash []byte // len = 32
	Location  *BlockLocation
	BlockHead block.Block
}

////////////////////////////////////////////////////
