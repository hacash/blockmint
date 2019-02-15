package store

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/types/block"
)

////////////////////////////////////////////////////

type BlockLocation struct {
	BlockFileNum uint32
	FileOffset   uint32
	BlockLen     uint32
}

func (loc *BlockLocation) Parse(buf []byte, seek uint32) error {
	loc.BlockFileNum = binary.BigEndian.Uint32(buf[0:4])
	loc.FileOffset = binary.BigEndian.Uint32(buf[4:8])
	loc.BlockLen = binary.BigEndian.Uint32(buf[8:12])
	return nil
}

func (loc *BlockLocation) Serialize() []byte {
	var byt1 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt1, loc.BlockFileNum)
	var byt2 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt2, loc.FileOffset)
	var byt3 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt3, loc.BlockLen)
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
