package store

import (
	"bytes"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/types/block"
)

var (
	ValueSizeSet = uint32(3*4 + 1 + 5 + 5 + 32 + 32 + 4)
)

type BlockIndexDB struct {
	filepath string

	treedb *hashtreedb.HashTreeDB
}

func (this *BlockIndexDB) Init(filepath string) {
	this.filepath = filepath
	this.treedb = hashtreedb.NewHashTreeDB(filepath, ValueSizeSet, 32)
	this.treedb.KeyReverse = true      // key值倒序
	this.treedb.FilePartitionLevel = 2 // 文件分区
}

func (this *BlockIndexDB) Save(hash []byte, blockLoc *BlockLocation, block block.Block) error {
	blockheadbytes, err := block.SerializeHead()
	if err != nil {
		return err
	}
	return this.SaveByBlockHeadByte(hash, blockLoc, blockheadbytes)
}

func (this *BlockIndexDB) SaveByBlockHeadByte(hash []byte, blockLoc *BlockLocation, blockheadbytes []byte) error {

	query, e := this.treedb.CreateQuery(hash)
	if e != nil {
		return e
	}
	// new body
	var bodybyte bytes.Buffer
	bodybyte.Write(blockLoc.Serialize())
	bodybyte.Write(blockheadbytes)
	// save
	e = query.Save(bodybyte.Bytes())
	if e != nil {
		return e
	}
	query.Close()
	// ok
	return nil
}

func (this *BlockIndexDB) Find(hash []byte) (*BlockLocation, block.Block, error) {
	query, e1 := this.treedb.CreateQuery(hash)
	if e1 != nil {
		return nil, nil, e1
	}
	result, e2 := query.Read()
	if e2 != nil {
		return nil, nil, e2
	}
	if uint32(len(result)) < ValueSizeSet {
		return nil, nil, err.New("file store error")
	}

	var loc BlockLocation
	loc.Parse(result, 0)
	var block, _, e3 = blocks.ParseBlockHead(result, 3*4)
	if e3 != nil {
		return nil, nil, e3
	}
	return &loc, block, nil
}

/*

func (db *BlockIndexDB) InitForTest(filename string) {
	db.filename = filename
	file, err := os.OpenFile(db.filename, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	if err != nil {
		fmt.Println(err)
	}
	db.indexfile = file
	defer func() {
		db.indexfile.Close()
		db.indexfile = nil
	}()
}

func (db *BlockIndexDB) Save(hash []byte, blockLoc *BlockLocation, block block.Block) error {
	return nil
}

func (db *BlockIndexDB) SaveForce(hash []byte, blockLoc *BlockLocation, block block.Block) error {
	blockheadbytes, _ := block.SerializeHead()
	return db.SaveByByteForce(hash, blockLoc, blockheadbytes)
}

func (db *BlockIndexDB) SaveByByteForce(hash []byte, blockLoc *BlockLocation, blockheadbytes []byte) error {

	db.indexfile, _ = os.OpenFile(db.filename, os.O_RDWR|os.O_CREATE, 0777)
	defer func() {
		db.indexfile.Close()
		db.indexfile = nil
	}()

	///////////

	filestat, e := db.indexfile.Stat()
	if e != nil {
		return e
	}
	var filesize = filestat.Size()
	// find
	var itempos = db.findItem(hash, true, false, false)
	if itempos == nil {
		//return err.New("SaveByByteForce error")
	}
	isReplaceCoverSet := false // 是否为覆盖旧的区块
	// new body
	var bodybyte bytes.Buffer
	bodybyte.Write(hash)
	bodybyte.Write(blockLoc.Serialize())
	//fmt.Println(blockheadbytes)
	bodybyte.Write(blockheadbytes)
	bodybyte.Write([]byte{0, 0}) // 空余
	// next ptr
	var valbyte = make([]byte, 4)
	writeNumber := filesize / int64(itemSizeSet)
	if filesize == 0 {
		writeNumber = 1 // 新文件
	}
	ptrNumber := writeNumber
	if itempos != nil && itempos.itemkind == 0 {
	} else if itempos != nil && itempos.itemkind == 2 {
		writeNumber += 1
	}
	binary.BigEndian.PutUint32(valbyte, uint32(writeNumber))
	var valkind = uint8(2)
	if itempos == nil {
	} else if itempos.itemkind == 0 {
	} else if itempos.itemkind == 2 {
		valkind = 1 // 新分叉
	}
	var nextptr = bytes.NewBuffer([]byte{valkind})
	nextptr.Write(valbyte)
	// new items
	var newitems []bytes.Buffer // 重复向下递归
	var newitem bytes.Buffer
	if itempos == nil || itempos.itemkind == 0 {
		var tailpos = 31
		if itempos != nil {
			tailpos = len(itempos.hashtail) - 1
		}
		var writeItemPos1 int
		var writeItemLi1 *bytes.Buffer
		var ipw = uint8(hash[tailpos]) % 25
		writeItemPos1 = int(ipw) * (itemLiWide)
		var valbyte = make([]byte, 4)
		binary.BigEndian.PutUint32(valbyte, uint32(writeNumber))
		writeItemLi1 = bytes.NewBuffer([]byte{2})
		writeItemLi1.Write(valbyte)
		newitem.Write(bytes.Repeat([]byte{0}, writeItemPos1))
		newitem.Write(writeItemLi1.Bytes())
		newitem.Write(bytes.Repeat([]byte{0}, itemSizeSet-len(newitem.Bytes())))
		newitems = append(newitems, newitem)
	} else if itempos.itemkind == 2 {

		// hash是否相同
		issame := bytes.Compare(hash, itempos.Blockhash) == 0
		if issame {
			// 同一个哈希， 覆盖之前的 区块值
			isReplaceCoverSet = true
		} else {
			// 新分叉
			var tailpos = len(itempos.hashtail) - 1
			db.createDownItem(&newitems, writeNumber+1, ptrNumber, itempos.itemnumber, itempos.Blockhash, hash, uint32(tailpos))
		}

	}
	// write 区块头数据一旦储存位置不会更改，只会追加新增数据和修改指针
	var filetailseek = filesize
	if filesize == 0 {
		//fmt.Println("filesize == 0")
		db.indexfile.WriteAt(newitems[0].Bytes(), 0)
		db.indexfile.WriteAt(bodybyte.Bytes(), int64(itemSizeSet))
	} else {
		if itempos == nil {
			fmt.Println("NOT FIND itempos")
			fmt.Println(filesize / 125)
			fmt.Println(hash)
		} else if itempos.itemkind == 0 {
			db.indexfile.WriteAt(bodybyte.Bytes(), filetailseek)
			db.indexfile.WriteAt(nextptr.Bytes(), itempos.itemoffset) // change ptr
		} else if itempos.itemkind == 2 {
			if isReplaceCoverSet {

				db.indexfile.WriteAt(bodybyte.Bytes(), int64(itempos.itemnumber*uint32(itemSizeSet))) // recover

			} else {

				db.indexfile.WriteAt(bodybyte.Bytes(), filetailseek)
				for i := 0; i < len(newitems); i++ { // 多个item
					db.indexfile.WriteAt(newitems[i].Bytes(), filetailseek+int64(i+1)*int64(itemSizeSet))
				}
				db.indexfile.WriteAt(nextptr.Bytes(), itempos.itemoffset) // change ptr

			}
		}
	}
	return nil
}

func (db *BlockIndexDB) createDownItem(totalitems *[]bytes.Buffer, tailnumber int64, newbodynumber int64, oldptrnumber uint32, oldhash []byte, hash []byte, tailseek uint32) {
	tailpos := tailseek
	writeNumber := tailnumber
	//fmt.Println(tailnumber)
	//fmt.Println(oldptrnumber)
	//fmt.Println(oldhash)
	//fmt.Println(hash)
	//fmt.Println(tailseek)
	var newitem bytes.Buffer
	//
	var ipw1 = uint8(hash[tailpos]) % 25
	writeItemPos1 := int(ipw1) * (itemLiWide)
	var ipw2 = uint8(oldhash[tailpos]) % 25
	writeItemPos2 := int(ipw2) * (itemLiWide)
	//fmt.Println(writeItemPos1)
	//fmt.Println(writeItemPos2)
	if writeItemPos2 != writeItemPos1 {
		// 到底
		var valbyte1 = make([]byte, 4)
		binary.BigEndian.PutUint32(valbyte1, uint32(newbodynumber))
		writeItemLi1 := bytes.NewBuffer([]byte{2})
		writeItemLi1.Write(valbyte1)
		// 旧区块
		//fmt.Println(writeItemPos2)
		var valbyte2 = make([]byte, 4)
		binary.BigEndian.PutUint32(valbyte2, uint32(oldptrnumber))
		writeItemLi2 := bytes.NewBuffer([]byte{2})
		writeItemLi2.Write(valbyte2)
		if writeItemPos2 < writeItemPos1 { // 调换
			writeItemPos1, writeItemPos2 = writeItemPos2, writeItemPos1
			writeItemLi1, writeItemLi2 = writeItemLi2, writeItemLi1
		}
		newitem.Write(bytes.Repeat([]byte{0}, writeItemPos1))
		newitem.Write(writeItemLi1.Bytes())
		newitem.Write(bytes.Repeat([]byte{0}, writeItemPos2-writeItemPos1-5))
		newitem.Write(writeItemLi2.Bytes())
		newitem.Write(bytes.Repeat([]byte{0}, itemSizeSet-len(newitem.Bytes())))
		total := append(*totalitems, newitem)
		*totalitems = *&total
	} else {

		// 前缀重复
		var valbyte1 = make([]byte, 4)
		binary.BigEndian.PutUint32(valbyte1, uint32(writeNumber))
		writeItemLi1 := bytes.NewBuffer([]byte{1})
		writeItemLi1.Write(valbyte1)

		newitem.Write(bytes.Repeat([]byte{0}, writeItemPos1))
		newitem.Write(writeItemLi1.Bytes())
		newitem.Write(bytes.Repeat([]byte{0}, itemSizeSet-len(newitem.Bytes())))
		// 写入
		total := append(*totalitems, newitem)
		*totalitems = *&total
		// 递归向下
		db.createDownItem(totalitems, tailnumber+1, newbodynumber, oldptrnumber, oldhash, hash, tailseek-1)

	}
}

func (db *BlockIndexDB) Find(hash []byte) *PositionFindItem {
	return db.findItem(hash, true, true, true)
}

func (db *BlockIndexDB) findItem(hash []byte, getitem bool, getlocation bool, getblockhead bool) *PositionFindItem {

	if db.indexfile == nil {
		file, ef := os.OpenFile(db.filename, os.O_RDWR, 0777)
		if ef != nil {
			return nil
		}
		db.indexfile = file
		defer func() {
			file.Close()
			db.indexfile = nil
		}()
	}

	///////////
	// 检查文件
	filestat, e := db.indexfile.Stat()
	if e != nil {
		return nil
	}
	var filesize = filestat.Size()
	if filesize == 0 {
		return nil
	}
	//fmt.Println(hash)
	// hash值 倒序
	hsdt := invertOrder(hash)
	//fmt.Println(hsdt)
	// 开始查找
	return db.recurrenceFindItem(0, hsdt, 0, getitem, getlocation, getblockhead)
}

func (db *BlockIndexDB) recurrenceFindItem(offset int64, hash []byte, seek uint32, getitem bool, getlocation bool, getblockhead bool) *PositionFindItem {

	var head = uint8(hash[seek]) % 25 // ==(itemSizeSet/itemLiWide)
	var readoffset = offset + int64(int(head)*itemLiWide)
	var idxbyte = make([]byte, itemLiWide)
	var hstail = hash[seek+1:]
	var hashtail = invertOrder(hstail)

	//fmt.Println(hash)

	rdlen, err := db.indexfile.ReadAt(idxbyte, readoffset)
	//fmt.Println(readoffset)
	//fmt.Println(idxbyte)
	if err != nil || rdlen != itemLiWide {
		return nil
	}
	var kind = uint8(idxbyte[0])
	if kind == 0 {
		return &PositionFindItem{
			itemkind:   0,
			itemoffset: readoffset,
			itemnumber: 0,
			hashtail:   hashtail,
		}
	}
	var itemnumber = binary.BigEndian.Uint32(idxbyte[1:5])
	var readitemstart = int64(itemnumber) * int64(itemSizeSet)
	//fmt.Println(itemnumber)
	//fmt.Println(readitemstart)
	if kind == 1 {
		// 跳转递归查询
		return db.recurrenceFindItem(readitemstart, hash, seek+1, getitem, getlocation, getblockhead)
	}
	if kind == 2 {
		item := &PositionFindItem{
			itemkind:   2,
			itemoffset: readoffset,
			itemnumber: itemnumber,
			hashtail:   hashtail,
		}
		if getitem {
			bodybyte := make([]byte, itemSizeSet)
			rdlen, err := db.indexfile.ReadAt(bodybyte, readitemstart)
			if err != nil || rdlen != itemSizeSet {
				return nil
			}
			//fmt.Println(bodybyte)
			//fmt.Println(bodybyte[0:32])
			//fmt.Println(bodybyte[32:32+3*4])
			//fmt.Println(bodybyte[32+3*4:])
			item.Blockhash = bodybyte[0:32]
			if getlocation {
				item.Location = new(BlockLocation)
				item.Location.Parse(bodybyte, 32)
			}
			if getblockhead {
				blk, _, err := blocks.ParseBlockHead(bodybyte, 32+3*4)
				if err != nil {
					return nil
				}
				item.BlockHead = blk
			}
		}
		return item
	}

	return nil

}

///////////////////////////////////////////

func invertOrder(hash []byte) []byte {
	var length = len(hash)
	var hsdt = make([]byte, length)
	copy(hsdt, hash)
	for i := 0; i < length/2; i++ {
		hsdt[i], hsdt[length-i-1] = hsdt[length-i-1], hsdt[i]
	}
	return hsdt
}

*/
