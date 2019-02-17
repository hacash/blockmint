package blockstore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/tests"
	"os"
	"testing"
)

func Test_store1(t *testing.T) {

	var testByteAry = tests.GenTestData_block()
	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)

	var testdir = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/blockstore/datas"
	//os.Remove(testdir)

	var db1 store.BlockIndexDB
	db1.Init(testdir)

	hashbyte, _ := hex.DecodeString("0000000000000000000258ed908d8ed4dffd87b87105e0b6e11a5e7f465b7400")
	loc := &store.BlockLocation{
		BlockFileNum: uint32(0),
		FileOffset:   uint32(0),
		DataLen:      uint32(0),
	}
	//fmt.Println(loc)
	db1.Save(hashbyte, loc, block1)

	result, blockhead, _ := db1.Find(hashbyte)
	fmt.Println(result.FileOffset)
	fmt.Println(blockhead.SerializeHead())

	//fmt.Println((256*3+1)*256*256*256*256 /1024/1024 )
}

func Test_store2(t *testing.T) {

	var testByteAry = tests.GenTestData_block()
	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)

	var testpath = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/blockstore/datas"
	os.Remove(testpath)

	var db1 store.BlockDataDB
	db1.Init(testpath)

	maxBlockNum := 10000
	for i := 0; i < maxBlockNum; i++ {
		one := make([]byte, 32)
		rand.Read(one)
		//fmt.Println(one)
		db1.SaveBlock(one, block1)
	}

	hashone1, _ := hex.DecodeString("FFFFFFFFFFFFFF01000000000000000000000000000000000000000000000001")
	fmt.Println(block1.Serialize())
	db1.SaveBlock(hashone1, block1)

	search, _ := db1.ReadBlock(hashone1)
	fmt.Println(search.Serialize())

}

func Test_store_loop(t *testing.T) {

	var testByteAry = tests.GenTestData_block()
	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)

	var testpath = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/blockstore/datas"
	os.Remove(testpath)

	var db1 store.BlockIndexDB
	db1.Init(testpath)

	loc := &store.BlockLocation{
		BlockFileNum: uint32(0),
		FileOffset:   uint32(0),
		DataLen:      uint32(0),
	}
	//fmt.Println(loc)

	/*
		maxBlockNum := 500000
		for i:=0; i<maxBlockNum-1; i++ {
			one := make([]byte, 32)
			rand.Read(one)
			//fmt.Println(one)
			db1.SaveForce(one, loc, block1)
		}
	*/
	hashone1, _ := hex.DecodeString("FFFFFFFFFFFFFF01000000000000000000000000000000000000000000000001")
	db1.Save(hashone1, loc, block1)
	hashone2, _ := hex.DecodeString("FFFFFFFFFFFFFF02000000000000000000000000000000000000000000000101")
	db1.Save(hashone2, loc, block1)
	hashone3, _ := hex.DecodeString("FFFFFFFFFFFFFF03000000000000000000000000000000000000000000010101")
	db1.Save(hashone3, loc, block1)
	hashone4, _ := hex.DecodeString("FFFFFFFFFFFFFF04000000000000000000000000000000000000000001010101")
	db1.Save(hashone4, loc, block1)
	hashone5, _ := hex.DecodeString("FFFFFFFFFFFFFF05000000000000000000000000000000000000000000000201")
	db1.Save(hashone5, loc, block1)
	hashone6, _ := hex.DecodeString("FFFFFFFFFFFFFF06000000000000000000000000000000000000000000000301")
	db1.Save(hashone6, loc, block1)
	hashone7, _ := hex.DecodeString("FFFFFFFFFFFFFF07000000000000000000000000000000000000000000010201")
	db1.Save(hashone7, loc, block1)
	hashone8, _ := hex.DecodeString("FFFFFFFFFFFFFF08000000000000000000000000000000000000000000010201")
	db1.Save(hashone8, loc, block1)
	hashone9, _ := hex.DecodeString("FFFFFFFFFFFFFF09000000000000000000000000000000000000000101010201")
	db1.Save(hashone9, loc, block1)
	hashone10, _ := hex.DecodeString("FFFFFFFFFFFFFF0a000000000000000000000000000000000000000201010201")
	db1.Save(hashone10, loc, block1)

	//maxBlockNum := 1000
	//for i := 0; i < maxBlockNum-1; i++ {
	//	db1.Find(hashone1)
	//	db1.Find(hashone2)
	//	db1.Find(hashone3)
	//	db1.Find(hashone4)
	//	db1.Find(hashone5)
	//	db1.Find(hashone6)
	//	db1.Find(hashone7)
	//	db1.Find(hashone8)
	//	db1.Find(hashone9)
	//	db1.Find(hashone10)
	//}

	//fmt.Println(res)

	//
	//var blk block.Block
	_, blk, _ := db1.Find(hashone1)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone2)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone3)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone4)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone5)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone6)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone7)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone8)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone9)
	fmt.Println(blk.Serialize())
	_, blk, _ = db1.Find(hashone10)
	fmt.Println(blk.Serialize())

	//fmt.Println((256*3+1)*256*256*256*256 /1024/1024 )
}
