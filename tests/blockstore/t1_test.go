package blockstore

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/tests"
	"testing"
)

func Test_store1(t *testing.T) {

	var testByteAry = tests.GenTestData_block()
	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)

	var db1 store.BlockIndexDB
	db1.InitForTest("/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/blockstore/indexdb.dat")

	hashbyte, _ := hex.DecodeString("0000000000000000000258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d")
	loc := &store.BlockLocation{
		BlockFileNum: uint32(0),
		FileOffset:   uint32(0),
		BlockLen:     uint32(0),
	}
	fmt.Println(loc)
	db1.SaveForce(hashbyte, loc, block1)
	db1.Save(hashbyte, loc, block1)

	result := db1.Find(hashbyte)

	fmt.Println(result)

	//fmt.Println((256*3+1)*256*256*256*256 /1024/1024 )
}

func Test_store_loop(t *testing.T) {

	var testByteAry = tests.GenTestData_block()
	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)

	var db1 store.BlockIndexDB
	db1.InitForTest("/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/blockstore/indexdb.dat")

	loc := &store.BlockLocation{
		BlockFileNum: uint32(0),
		FileOffset:   uint32(0),
		BlockLen:     uint32(0),
	}
	fmt.Println(loc)

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
	//db1.SaveForce(hashone1, loc, block1)
	hashone2, _ := hex.DecodeString("FFFFFFFFFFFFFF02000000000000000000000000000000000000000000000101")
	//db1.SaveForce(hashone2, loc, block1)
	hashone3, _ := hex.DecodeString("FFFFFFFFFFFFFF03000000000000000000000000000000000000000000010101")
	//db1.SaveForce(hashone3, loc, block1)
	hashone4, _ := hex.DecodeString("FFFFFFFFFFFFFF04000000000000000000000000000000000000000001010101")
	//db1.SaveForce(hashone4, loc, block1)
	hashone5, _ := hex.DecodeString("FFFFFFFFFFFFFF05000000000000000000000000000000000000000000000201")
	//db1.SaveForce(hashone5, loc, block1)
	hashone6, _ := hex.DecodeString("FFFFFFFFFFFFFF06000000000000000000000000000000000000000000000301")
	//db1.SaveForce(hashone6, loc, block1)
	hashone7, _ := hex.DecodeString("FFFFFFFFFFFFFF07000000000000000000000000000000000000000000010201")
	//db1.SaveForce(hashone7, loc, block1)
	hashone8, _ := hex.DecodeString("FFFFFFFFFFFFFF08000000000000000000000000000000000000000000010201")
	//db1.SaveForce(hashone8, loc, block1)
	hashone9, _ := hex.DecodeString("FFFFFFFFFFFFFF09000000000000000000000000000000000000000101010201")
	//db1.SaveForce(hashone9, loc, block1)
	hashone10, _ := hex.DecodeString("FFFFFFFFFFFFFF10000000000000000000000000000000000000000201010201")
	//db1.SaveForce(hashone10, loc, block1)
	db1.Save(hashone1, loc, block1)

	maxBlockNum := 1000
	for i := 0; i < maxBlockNum-1; i++ {
		db1.Find(hashone1)
		db1.Find(hashone2)
		db1.Find(hashone3)
		db1.Find(hashone4)
		db1.Find(hashone5)
		db1.Find(hashone6)
		db1.Find(hashone7)
		db1.Find(hashone8)
		db1.Find(hashone9)
		db1.Find(hashone10)
	}

	//fmt.Println(res)

	//
	//
	//fmt.Println(db1.Find(hashone1).Blockhash)
	//fmt.Println(db1.Find(hashone2).Blockhash)
	//fmt.Println(db1.Find(hashone3).Blockhash)
	//fmt.Println(db1.Find(hashone4).Blockhash)
	//fmt.Println(db1.Find(hashone5).Blockhash)
	//fmt.Println(db1.Find(hashone6).Blockhash)
	//fmt.Println(db1.Find(hashone7).Blockhash)
	//fmt.Println(db1.Find(hashone8).Blockhash)
	//fmt.Println(db1.Find(hashone9).Blockhash)
	//fmt.Println(db1.Find(hashone10).Blockhash)

	//fmt.Println((256*3+1)*256*256*256*256 /1024/1024 )
}
