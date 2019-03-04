package hashtreedb

import (
	"bytes"
	"encoding/binary"
)

var IndexItemSize = uint32(5) // 索引项宽度

type IndexItem struct {
	ItemFindOffset int64 // 数据位置数

	Type        uint8  // 0:nil 1:枝 2:叶
	ValuePtrNum uint32 // 指针位置数 ×seg=offset

	searchLevel uint32 // 搜索次数，从0开始

	ItemHash  []byte
	ValueBody []byte
}

func NewIndexItem(ty uint8, ptrNum uint32) *IndexItem {
	return &IndexItem{
		Type:        ty,
		ValuePtrNum: ptrNum,
	}
}

func (this *IndexItem) Parse(buf []byte, seek uint32) error {
	this.Type = uint8(buf[seek])
	this.ValuePtrNum = binary.BigEndian.Uint32(buf[seek+1 : seek+5])
	return nil
}

func (this *IndexItem) Serialize() []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte{this.Type})
	var byt1 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt1, this.ValuePtrNum)
	buffer.Write(byt1)
	return buffer.Bytes()
}
