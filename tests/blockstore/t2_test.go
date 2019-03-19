package blockstore

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/tests"
	"os"
	"testing"
)

func Test_gc(t *testing.T) {

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
	fmt.Println(loc)

	hashone1, _ := hex.DecodeString("FFFFFFFFFFFFFF01000000000000000000000000000000000000000000000001")
	//db1.Save(hashone1, loc, block1)
	//hashone2, _ := hex.DecodeString("FFFFFFFFFFFFFF02000000000000000000000000000000000000000000000101")
	//db1.Save(hashone2, loc, block1)
	//hashone3, _ := hex.DecodeString("FFFFFFFFFFFFFF03000000000000000000000000000000000000000000010101")
	//db1.Save(hashone3, loc, block1)
	//hashone4, _ := hex.DecodeString("FFFFFFFFFFFFFF04000000000000000000000000000000000000000001010101")
	//db1.Save(hashone4, loc, block1)
	//hashone5, _ := hex.DecodeString("FFFFFFFFFFFFFF05000000000000000000000000000000000000000000000201")
	//db1.Save(hashone5, loc, block1)
	//hashone6, _ := hex.DecodeString("FFFFFFFFFFFFFF06000000000000000000000000000000000000000000000301")
	//db1.Save(hashone6, loc, block1)
	//hashone7, _ := hex.DecodeString("FFFFFFFFFFFFFF07000000000000000000000000000000000000000000010201")
	//db1.Save(hashone7, loc, block1)
	//hashone8, _ := hex.DecodeString("FFFFFFFFFFFFFF08000000000000000000000000000000000000000000010201")
	//db1.Save(hashone8, loc, block1)
	//hashone9, _ := hex.DecodeString("FFFFFFFFFFFFFF09000000000000000000000000000000000000000101010201")
	//db1.Save(hashone9, loc, block1)
	//hashone10, _ := hex.DecodeString("FFFFFFFFFFFFFF0a000000000000000000000000000000000000000201010201")
	//db1.Save(hashone10, loc, block1)

}


