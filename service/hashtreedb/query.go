package hashtreedb

import (
	"bytes"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/sys/util"
	"os"
	"strconv"
	"sync"
)

// 查询实例

type QueryInstance struct {
	db *HashTreeDB

	operationHash []byte // 要操作的哈希
	queryKeyHash  []byte // 去掉前缀文件分区，展开精度的key值
	wideKeyHash   []byte // 包含前缀的key值

	targetFile     *os.File // 当前正在使用的文件
	targetFileName *string  // 当前正在使用的文件

}

func NewQueryInstance(db *HashTreeDB, hash []byte, keyhash []byte, querykey []byte, file *os.File, filename *string) *QueryInstance {
	// 展开数据，避免损失精度

	ins := &QueryInstance{
		db,
		hash,
		querykey,
		keyhash,
		file,
		filename,
	}

	//fmt.Println( "db.MenuWide", db.MenuWide )
	//fmt.Println( "old key", key)
	//fmt.Println( "queryKey ---- ", ins.queryKeyHash)

	return ins
}

// 关闭
func (this *QueryInstance) Close() error {
	defer func() {
		//fmt.Println(*this.targetFileName)
		if lock, has := this.db.FileLock.Load(*this.targetFileName); has {
			//fmt.Println("Unlock file " + *this.targetFileName)
			this.db.FileLock.Delete(*this.targetFileName)
			lock.(*sync.Mutex).Unlock()
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

// 更新 item 索引数据
func (this *QueryInstance) UpdateItem(item *IndexItem) (*IndexItem, error) {
	_, e := this.targetFile.WriteAt(item.Serialize(), item.ItemFindOffset)
	return item, e
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
	var realbodybuf bytes.Buffer
	realbodybuf.Write(this.operationHash)
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
			gcmng, e1 := this.db.GetGcService(this.operationHash)
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
		oldqueryhashkey := this.db.GetQueryHashKey(oldhash)
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
			//fmt.Println("--------------")
			//fmt.Println(oldhash)
			//fmt.Println(this.operationHash)
			//fmt.Println(oldqueryhashkey)
			//fmt.Println(this.queryKeyHash)
			//fmt.Println("--------------")
			for i := item.searchLevel + 1; i < uint32(len(oldqueryhashkey)) || i < uint32(len(this.queryKeyHash)); i++ {
				//fmt.Println("------", i, len(oldqueryhashkey))
				headOld := uint8(0)
				headInsert := uint8(0)
				if i < uint32(len(oldqueryhashkey)) {
					headOld = oldqueryhashkey[i]
				}
				if i < uint32(len(this.queryKeyHash)) {
					headInsert = this.queryKeyHash[i]
				}
				// 到达不同位置：key长度不一样，短的用0补齐
				//fmt.Println(headOld, headInsert)
				pos1 := uint32(headOld % this.db.MenuWide)
				pos2 := uint32(headInsert % this.db.MenuWide)
				//fmt.Println(pos1, pos2)
				var onebytes = make([]byte, segmentsize)
				if headOld == headInsert { // 向下
					writeCurrentOffsetNum += 1 // 向后推 1 segment
					itempath := NewIndexItem(1, uint32(writeCurrentOffsetNum))
					util.BytesCopyCover(onebytes, itempath.Serialize(), int(pos1*IndexItemSize))
					fileWriteAppendBatch.Write(onebytes)
					continue
				} else { // 分支
					writeCurrentOffsetNum += 1 // 向后推 1 segment
					//fmt.Println("+++++++++++++++")
					//fmt.Println(headOld, headInsert)
					//fmt.Println(pos1, pos2)
					//fmt.Println(item.ValuePtrNum, uint32(writeCurrentOffsetNum))
					//fmt.Println("+++++++++++++++")
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
		//fmt.Println("realvaluebody:")
		//fmt.Println(realvaluebody)
		//fmt.Println(hex.EncodeToString(realvaluebody))
		//fmt.Println(len(realvaluebody))
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
	for i := 0; ; i++ {
		posnum := uint8(0)
		if i < len(this.queryKeyHash) {
			posnum = this.queryKeyHash[i]
		} else {
			//fmt.Println(" FindItem add tail 0 + ")
		}
		headpos := uint32(posnum % this.db.MenuWide)
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
		//fmt.Println(" it := item.Type ", posnum, it)
		if it == 0 {
			return &item, nil
		} else if it == 2 {
			if getvalue {
				body := make([]byte, segmentsize)
				rdlen, e := this.targetFile.ReadAt(body, int64(item.ValuePtrNum)*int64(segmentsize))
				if e != nil {
					if e.Error() == "EOF" && rdlen == 0 {
						return nil, nil
					}
					//fmt.Println(e.Error())
					//fmt.Println("-----------====================++++++++++++++++++++++++", e, e.Error(), body)
					return nil, e
				}
				item.ItemHash = body[0:this.db.HashSize]
				item.ValueBody = body[this.db.HashSize:]
			}
			//if item.searchLevel > 4 {
			//	fmt.Println(this.queryKeyHash)
			//	fmt.Println("item.searchLevel ", item.searchLevel)
			//}
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
	if item != nil {
		// delete mark
		if this.db.DeleteMark {
			// 清空数据内容而不删除key， 表示删除标记，用于db在copy cover 时删除目标数据库的条目
			item.Type = 3 // 删除标记
			this.UpdateItem(item) // 更新item索引数据，写入删除标记
			//this.Write(item, []byte{}) // mark // 写入空byte表示清空 segment， 表示删除
			return nil
		}else if item.Type == 2 {
			// fmt.Println(item.ItemHash)
			if 0 == bytes.Compare(this.operationHash, item.ItemHash) {
				return this.Delete(item) // 彻底删除数据
			}
		}
	}
	return nil
}

// 删除
func (this *QueryInstance) Delete(item *IndexItem) error {
	// 重置数据条目空间【测试】
	segsize := this.db.getSegmentSize()
	this.targetFile.WriteAt(bytes.Repeat([]byte{0}, int(segsize)), int64(item.ValuePtrNum)*int64(segsize))
	// 垃圾收集
	if this.db.OpenGc {
		// 使用gc
		gcmng, e1 := this.db.GetGcService(this.operationHash)
		if e1 == nil && gcmng != nil {
			gcmng.Collect(item.ValuePtrNum) // gc collect // 垃圾空间收集
		}
		item.Type = uint8(0)
		item.ValuePtrNum = 0
	} else {
		item.Type = uint8(3) // type = delete
	}
	// 修改指针删除
	_, e := this.UpdateItem(item)
	if e != nil {
		return e
	}
	return nil
}

/////////////////////////////////////////////////////
