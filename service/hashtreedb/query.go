package hashtreedb

import (
	"bytes"
	"github.com/hacash/blockmint/sys/err"
	"os"
	"strconv"
)

// 查询实例

type QueryInstance struct {
	db *HashTreeDB

	keyHash       []byte // key值
	operationHash []byte // 要操作的哈希

	targetFile *os.File // 当前正在使用的文件

}

func NewQueryInstance(db *HashTreeDB, hash []byte, key []byte, file *os.File) *QueryInstance {
	return &QueryInstance{
		db,
		key,
		hash,
		file,
	}
}

// 关闭
func (this *QueryInstance) Close() error {
	return this.targetFile.Close()
}

// 储存
func (this *QueryInstance) Save(value []byte) error {
	item, e1 := this.FindItem(true)
	if e1 != nil {
		return e1
	}
	return this.Write(item, value)
}

// 储存写入（hash相同会覆盖）
func (this *QueryInstance) Write(item *IndexItem, valuebody []byte) error {
	bodylen := uint32(len(valuebody))
	segmentsize := this.db.getSegmentSize()
	hxsize := this.db.HashSize
	if bodylen+hxsize > segmentsize {
		return err.New("Value bytes size overflow, max is" + strconv.Itoa(int(segmentsize-hxsize)))
	}
	filestat, e2 := this.targetFile.Stat()
	if e2 != nil {
		return e2
	}
	filesize := filestat.Size()
	if filesize > 0 && item == nil { // item == nil
		return err.New("file read error")
	}
	if filesize == 0 {
		_, e3 := this.targetFile.Write(bytes.Repeat([]byte{0}, int(segmentsize))) // init file
		if e3 != nil {
			return e3
		}
		filesize = int64(segmentsize)
	}
	// offset
	writeUpdateItemOffset := int64(this.keyHash[0])
	if item != nil {
		writeUpdateItemOffset = item.PositionOffset
	}
	// level index
	var fileWriteAppendBatch bytes.Buffer
	var itemIndexType = uint8(2)
	// pos
	if item != nil || item.Type == 2 {
		writeCurrentOffsetNum := filesize / int64(segmentsize)
		// 循环向下比对
		oldhash := item.ValueBody[0:hxsize]
		oldhashkey := this.db.getHashKey(oldhash)
		for i := item.searchLevel + 1; i < uint32(hxsize)-item.searchLevel; i++ {
			headOld := oldhashkey[i]
			headInsert := this.keyHash[i]
			pos1 := uint32(headOld) % this.db.MenuWide
			pos2 := uint32(headInsert) % this.db.MenuWide
			var onebytes = make([]byte, segmentsize)
			if headOld == headInsert { // 向下
				itempath := NewIndexItem(1, uint32(writeCurrentOffsetNum))
				copyCoverBytes(onebytes, itempath.Serialize(), int(pos1*IndexItemSize))
			} else { // 分支
				item1 := NewIndexItem(2, item.ValuePtrNum) // old
				copyCoverBytes(onebytes, item1.Serialize(), int(pos1*IndexItemSize))
				item2 := NewIndexItem(2, uint32(writeCurrentOffsetNum))
				copyCoverBytes(onebytes, item2.Serialize(), int(pos2*IndexItemSize))
			}
			writeCurrentOffsetNum += 1 // 向后推 1 segment
			fileWriteAppendBatch.Write(onebytes)
		}
	}
	// write file
	fileWriteAppendBatch.Write(valuebody)
	_, e := this.targetFile.WriteAt(fileWriteAppendBatch.Bytes(), filesize) // append value and indexs
	if e != nil {
		return e
	}
	// update item
	itemup := NewIndexItem(itemIndexType, uint32(filesize/int64(segmentsize)))
	_, e = this.targetFile.WriteAt(itemup.Serialize(), writeUpdateItemOffset) // update ptr
	if e != nil {
		return e
	}
	// ok
	return nil
}

// 查询
func (this *QueryInstance) FindItem(getvalue bool) (*IndexItem, error) {
	var fileoffset = int64(0)
	for i := 0; i < len(this.keyHash); i++ {
		headpos := uint32(this.keyHash[i]) % this.db.MenuWide
		readseek := fileoffset + int64(headpos)*int64(IndexItemSize)
		itembytes := make([]byte, IndexItemSize)
		rdlen, e := this.targetFile.ReadAt(itembytes, readseek)
		if e != nil {
			return nil, e
		}
		if rdlen == 0 {
			return nil, nil
		} else if uint32(rdlen) != IndexItemSize {
			return nil, err.New("file end")
		}
		var item IndexItem
		item.Parse(itembytes, 0)
		item.PositionOffset = readseek
		item.searchLevel = uint32(i) // 查询次数
		it := itembytes[0]
		if it == 0 {
			return &item, nil
		} else if it == 2 {
			if getvalue {
				segmentsize := this.db.getSegmentSize()
				valuebody := make([]byte, segmentsize)
				_, e := this.targetFile.ReadAt(valuebody, int64(item.ValuePtrNum)*int64(segmentsize))
				if e != nil {
					return nil, e
				}
				item.ValueBody = valuebody
			}
			return &item, nil
		} else if it == 1 {
			// 继续查询
			fileoffset = int64(item.ValuePtrNum) * int64(IndexItemSize)
		}
	}
	return nil, nil
}

// 读取
func (this *QueryInstance) Read() ([]byte, error) {

	item, e1 := this.FindItem(true)
	if e1 != nil {
		return nil, e1 // error
	}
	if item.Type == 0 {
		return []byte{}, nil // not find
	}
	if item.Type == 2 {
		itemhash := item.ValueBody[0:this.db.HashSize]
		if 0 == bytes.Compare(this.operationHash, itemhash) {
			return item.ValueBody[this.db.HashSize:], nil
		} else {
			return []byte{}, nil // not find
		}
	}
	return nil, err.New("Not find item type")
}

/////////////////////////////////////////////////////
