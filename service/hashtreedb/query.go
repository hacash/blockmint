package hashtreedb

import (
	"bytes"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/sys/util"
	"os"
	"strconv"
)

// 查询实例

type QueryInstance struct {
	db *HashTreeDB

	queryKeyHash  []byte // key值
	wideKeyHash   []byte // key值
	operationHash []byte // 要操作的哈希

	targetFile     *os.File // 当前正在使用的文件
	targetFileName *string  // 当前正在使用的文件

}

func NewQueryInstance(db *HashTreeDB, hash []byte, keyhash []byte, key []byte, file *os.File, filename *string) *QueryInstance {
	return &QueryInstance{
		db,
		key,
		keyhash,
		hash,
		file,
		filename,
	}
}

// 关闭
func (this *QueryInstance) Close() error {
	defer func() {
		if lock, has := this.db.FileLock[*this.targetFileName]; has {
			//fmt.Println("Unlock file " + *this.targetFileName)
			lock.Unlock()
		}
	}()
	return this.targetFile.Close()
}

// 储存
func (this *QueryInstance) Save(value []byte) (*IndexItem, error) {
	item, e1 := this.FindItem(true)
	if e1 != nil {
		return nil, e1
	}
	return this.Write(item, value)
}

// 储存写入（hash相同会覆盖）
func (this *QueryInstance) Write(item *IndexItem, valuebody []byte) (*IndexItem, error) {
	bodylen := uint32(len(valuebody))
	segmentsize := this.db.getSegmentSize()
	hxsize := this.db.HashSize
	overplussize := segmentsize - (bodylen + hxsize)
	if overplussize < 0 {
		return nil, err.New("Value bytes size overflow, max is" + strconv.Itoa(int(segmentsize-hxsize)))
	}
	realbodybuf := bytes.NewBuffer(this.operationHash)
	realbodybuf.Write(valuebody)
	if overplussize > 0 {
		realbodybuf.Write(bytes.Repeat([]byte{0}, int(overplussize)))
	}
	realvaluebody := realbodybuf.Bytes()

	filestat, e2 := this.targetFile.Stat()
	if e2 != nil {
		return nil, e2
	}
	filesize := filestat.Size()
	if filesize > 0 && item == nil { // item == nil
		return nil, err.New("file read error")
	}
	if filesize == 0 {
		_, e3 := this.targetFile.Write(bytes.Repeat([]byte{0}, int(segmentsize))) // init file
		if e3 != nil {
			return nil, e3
		}
		filesize = int64(segmentsize)
	}
	// offset
	writeUpdateItemOffset := (int64(this.queryKeyHash[0]) % int64(this.db.MenuWide)) * int64(IndexItemSize)
	if item != nil {
		writeUpdateItemOffset = item.ItemFindOffset
	}
	// level index
	var fileWriteAppendBatch bytes.Buffer
	var itemIndexType = uint8(2)
	var returnIndexItem *IndexItem = nil
	// gc
	var hasUseGC bool = false
	var gcPtrNum = uint32(0)
	// pos  0:null 3:delete
	if item != nil && item.Type == 0 {
		// 使用gc
		if this.db.OpenGc {
			gcmng, e1 := this.db.GetGcService(this.wideKeyHash)
			if gcmng != nil && e1 == nil {
				gcptr, ok, e := gcmng.Release()
				if gcptr > 0 && ok && e == nil {
					gcPtrNum = gcptr
					hasUseGC = true
				}
			}
		}
	}
	if item != nil && item.Type == 3 && item.ValuePtrNum > 0 {
		gcPtrNum = item.ValuePtrNum
		hasUseGC = true
	}
	if item != nil && item.Type == 2 {
		writeCurrentOffsetNum := filesize / int64(segmentsize)
		// 循环向下比对
		oldhash := item.ItemHash
		oldhashkey := this.db.GetQueryHashKey(oldhash)
		if 0 == bytes.Compare(oldhash, this.operationHash) {
			// 覆盖储存
			_, e := this.targetFile.WriteAt(realvaluebody, int64(item.ValuePtrNum)*int64(segmentsize)) // update ptr
			if e != nil {
				return nil, e
			}
			// ok
			return item, nil // isRecoverValue = true

		} else {
			// 仅key前缀相同
			itemIndexType = 1
			//fmt.Println(oldhashkey)
			//fmt.Println(this.queryKeyHash)
			for i := item.searchLevel + 1; i < uint32(hxsize)-item.searchLevel && i < uint32(len(oldhashkey)) && i < uint32(len(this.queryKeyHash)); i++ {
				//fmt.Println("------", i, len(oldhashkey))
				headOld := oldhashkey[i]
				headInsert := this.queryKeyHash[i]
				//fmt.Println(headOld, headInsert)
				pos1 := uint32(headOld) % this.db.MenuWide
				pos2 := uint32(headInsert) % this.db.MenuWide
				var onebytes = make([]byte, segmentsize)
				if headOld == headInsert { // 向下
					writeCurrentOffsetNum += 1 // 向后推 1 segment
					itempath := NewIndexItem(1, uint32(writeCurrentOffsetNum))
					util.BytesCopyCover(onebytes, itempath.Serialize(), int(pos1*IndexItemSize))
					fileWriteAppendBatch.Write(onebytes)
					continue
				} else { // 分支
					writeCurrentOffsetNum += 1 // 向后推 1 segment
					//fmt.Println(pos1)
					//fmt.Println(pos2)
					item1 := NewIndexItem(2, item.ValuePtrNum) // old
					util.BytesCopyCover(onebytes, item1.Serialize(), int(pos1*IndexItemSize))
					item2 := NewIndexItem(2, uint32(writeCurrentOffsetNum))
					util.BytesCopyCover(onebytes, item2.Serialize(), int(pos2*IndexItemSize))
					returnIndexItem = item2
					//fmt.Println(item1.Serialize())
					//fmt.Println(item2.Serialize())
					//fmt.Println(onebytes)
					fileWriteAppendBatch.Write(onebytes)
					break
				}
			}
		}
	}
	// write file
	if !hasUseGC {
		fileWriteAppendBatch.Write(realvaluebody) // 没有采用gc空间时，追加写入value
	}
	realwriteappend := fileWriteAppendBatch.Bytes()
	// fmt.Println(realwriteappend)
	if len(realwriteappend) > 0 {
		_, e := this.targetFile.WriteAt(realwriteappend, filesize) // append value and indexs
		if e != nil {
			return nil, e
		}
	}
	// update item
	itemupptrnum := uint32(filesize / int64(segmentsize))
	if hasUseGC {
		itemupptrnum = gcPtrNum // 采用gc空间，覆盖写入
		_, e := this.targetFile.WriteAt(realvaluebody, int64(gcPtrNum)*int64(segmentsize))
		if e != nil {
			return nil, e
		}
	}
	if bodylen == 0 {
		itemIndexType = 3 // delete mark
	}
	// update ptr
	itemup := NewIndexItem(itemIndexType, itemupptrnum)
	_, e := this.targetFile.WriteAt(itemup.Serialize(), writeUpdateItemOffset) // update ptr
	if e != nil {
		return nil, e
	}
	if returnIndexItem == nil {
		returnIndexItem = itemup
	}
	// ok
	return returnIndexItem, nil
}

// 查询
func (this *QueryInstance) FindItem(getvalue bool) (*IndexItem, error) {
	var segmentsize = this.db.getSegmentSize()
	var fileoffset = int64(0)
	for i := 0; i < len(this.queryKeyHash); i++ {
		headpos := uint32(this.queryKeyHash[i]) % this.db.MenuWide
		readseek := fileoffset + int64(headpos)*int64(IndexItemSize)
		itembytes := make([]byte, IndexItemSize)
		rdlen, e := this.targetFile.ReadAt(itembytes, readseek)
		//fmt.Println(itembytes)
		if e != nil {
			if e.Error() == "EOF" && rdlen == 0 {
				return nil, nil
			}
			return nil, e
		}
		if uint32(rdlen) != IndexItemSize {
			return nil, err.New("file end")
		}

		var item IndexItem
		item.Parse(itembytes, 0)
		item.ItemFindOffset = readseek
		item.searchLevel = uint32(i) // 查询次数
		it := item.Type
		if it == 0 {
			return &item, nil
		} else if it == 2 {
			if getvalue {
				body := make([]byte, segmentsize)
				_, e := this.targetFile.ReadAt(body, int64(item.ValuePtrNum)*int64(segmentsize))
				if e != nil {
					//fmt.Println(e.Error())
					return nil, e
				}
				item.ItemHash = body[0:this.db.HashSize]
				item.ValueBody = body[this.db.HashSize:]
			}
			return &item, nil
		} else if it == 1 {
			// 继续查询
			fileoffset = int64(item.ValuePtrNum) * int64(segmentsize)
		}
	}
	return nil, nil
}

// 读取
func (this *QueryInstance) Read() ([]byte, *IndexItem, error) {

	item, e1 := this.FindItem(true)
	if e1 != nil {
		return nil, nil, e1 // error
	}
	if item == nil {
		return nil, nil, nil // not find
	}
	if item.Type == 0 {
		return nil, nil, nil // not find
	}
	if item.Type == 2 {
		// fmt.Println(item.ItemHash)
		if 0 == bytes.Compare(this.operationHash, item.ItemHash) {
			return item.ValueBody, item, nil
		} else {
			return nil, nil, nil // not find
		}
	}
	return nil, nil, err.New("Not find item type")
}

// 删除
func (this *QueryInstance) Remove() error {

	item, e1 := this.FindItem(true)
	if e1 != nil {
		return e1 // error
	}
	if item != nil && item.Type == 2 {
		// fmt.Println(item.ItemHash)
		if 0 == bytes.Compare(this.operationHash, item.ItemHash) {
			return this.Delete(item) // 删除
		}
	}
	// delete mark
	if item != nil && this.db.DeleteMark {
		this.Write(item, []byte{}) // mark
	}
	return nil
}

// 删除
func (this *QueryInstance) Delete(item *IndexItem) error {
	tarptr := item.ValuePtrNum
	if this.db.OpenGc {
		// 使用gc
		gcmng, e1 := this.db.GetGcService(this.wideKeyHash)
		if e1 == nil && gcmng != nil {
			gcmng.Collect(tarptr) // gc collect
		}
		item.Type = uint8(0)
		item.ValuePtrNum = 0
	} else {
		item.Type = uint8(3) // type = delete
	}
	// 修改指针删除
	_, e := this.targetFile.WriteAt(item.Serialize(), item.ItemFindOffset)
	if e != nil {
		return e
	}
	return nil
}

/////////////////////////////////////////////////////
