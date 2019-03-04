package block

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/tests"
	"testing"
)

func Test_block_save_by_hight(t *testing.T) {

	// config
	config.LoadConfigFile()

	rtbytes := tests.GenTestData_block()
	rootblk, _, _ := blocks.ParseBlock(rtbytes, 0)
	db := store.GetGlobalInstanceBlocksDataStore()
	db.Save(rootblk)

	fmt.Println(rootblk.Hash())
	fmt.Println(hex.EncodeToString(rtbytes))

	blkbytes, e := db.GetBlockBytesByHeight(rootblk.GetHeight(), true, true)
	if e != nil {
		panic(e)
	}
	fmt.Println(hex.EncodeToString(blkbytes))

	bbb := make([]byte, len(blkbytes))
	copy(bbb, blkbytes)
	blk2, _, _ := blocks.ParseBlock(bbb, 0)

	fmt.Println(blk2.Hash())

}

func Test_block_save_by_height(t *testing.T) {

	// config
	config.LoadConfigFile()

	for i := uint32(0); ; i++ {

		rtbytes := tests.GenTestData_block_set_height(i + 1)
		rootblk, _, _ := blocks.ParseBlock(rtbytes, 0)
		db := store.GetGlobalInstanceBlocksDataStore()
		db.Save(rootblk)

		//fmt.Println(rootblk.Hash())
		//fmt.Println(hex.EncodeToString( rtbytes ))

		blkbytes, e := db.GetBlockBytesByHeight(rootblk.GetHeight(), true, true)
		if e != nil {
			panic(e)
		}
		if len(blkbytes) == 0 {
			panic(rootblk.GetHeight())
		}
		//fmt.Println(i)
		//if i % 5000 == 0 {
		fmt.Println(i, hex.EncodeToString(blkbytes[0:30]))
		//}

		bbb := make([]byte, len(blkbytes))
		copy(bbb, blkbytes)
		//fmt.Println(hex.EncodeToString( bbb ))
		//blk2, _, _ := blocks.ParseBlock(bbb, 0)
		//fmt.Println(blk2.Hash())

	}

}
