package store

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/service/hashtreedb"
	"github.com/hacash/blockmint/sys/err"
)

// Height 区块高度索引库

var (
	valueSizeSetBlockHeightDB = BlockLocationSize + 2 + 4
)

type BlockHeightDBItemData struct {
	BlockHeadInfoFilePartition [2]byte        // 区块头信息文件分区标号
	BlockHeadInfoPtrNumber     uint32         // 区块头信息指针位置
	location                   *BlockLocation // 区块文件中的位置
}

func (this *BlockHeightDBItemData) Parse(buf []byte, seek uint32) error {
	this.BlockHeadInfoFilePartition = [2]byte{buf[seek], buf[seek+1]}
	this.BlockHeadInfoPtrNumber = binary.BigEndian.Uint32(buf[seek+2 : seek+6])
	var loc BlockLocation
	loc.Parse(buf, seek+6)
	this.location = &loc
	return nil
}

func (this *BlockHeightDBItemData) Serialize() []byte {
	var buffer bytes.Buffer
	buffer.Write(this.BlockHeadInfoFilePartition[:])
	var byt1 = make([]byte, 4)
	binary.BigEndian.PutUint32(byt1, this.BlockHeadInfoPtrNumber)
	buffer.Write(byt1)
	buffer.Write(this.location.Serialize())
	return buffer.Bytes()
}

/////////////////////////////////////

type BlockHeightDB struct {
	filepath string
	treedb   *hashtreedb.HashTreeDB
}

func (this *BlockHeightDB) Init(filepath string) {
	this.filepath = filepath
	this.treedb = hashtreedb.NewHashTreeDB(filepath, valueSizeSetBlockHeightDB, 7)
	this.treedb.FilePartitionLevel = 3 // 文件分区
}

func (this *BlockHeightDB) Save(height uint64, partnum [2]byte, headptrnum uint32, blkloc *BlockLocation) error {

	//fmt.Println("this.treedb.CreateQuery(key) ", key)

	query, e := this.treedb.CreateQuery(this.dealKey(height))
	defer query.Close()
	if e != nil {
		return e
	}
	// save
	item := BlockHeightDBItemData{
		BlockHeadInfoFilePartition: partnum,
		BlockHeadInfoPtrNumber:     headptrnum,
		location:                   blkloc,
	}
	valbytes := item.Serialize()
	//fmt.Println("valbytes ", valbytes)
	_, e1 := query.Save(valbytes)
	if e1 != nil {
		return e1
	}
	// ok
	return nil

}

func (this *BlockHeightDB) Find(height uint64) (*BlockHeightDBItemData, error) {

	query, e1 := this.treedb.CreateQuery(this.dealKey(height))
	defer query.Close()
	if e1 != nil {
		return nil, e1
	}
	// read
	result, _, e2 := query.Read()
	//fmt.Println("result ", result)
	if e2 != nil {
		return nil, e2
	}
	rdlen := uint32(len(result))
	if rdlen == 0 {
		return nil, nil // empty file
	}
	if rdlen < trsIdxItemSizeSet {
		return nil, err.New("file store error")
	}
	var item BlockHeightDBItemData
	item.Parse(result, 0)
	if item.location == nil {
		return nil, nil
	}

	return &item, nil
}

// key length = 12
func (this *BlockHeightDB) dealKey(height uint64) []byte {
	/*
		var key1 = make([]byte, 4) // 小端保存key
		binary.LittleEndian.PutUint32(key1, uint32(height))
		var key2 = make([]byte, 8) // 小端保存key
		binary.LittleEndian.PutUint64(key2, height)
		var buf bytes.Buffer
		buf.Write( key1 )
		buf.Write( key2 )
		//fmt.Println( buf.Bytes() )
		return buf.Bytes()
	*/
	var key2 = make([]byte, 8) // 小端保存key
	binary.LittleEndian.PutUint64(key2, height)
	return key2[0:7]
}
