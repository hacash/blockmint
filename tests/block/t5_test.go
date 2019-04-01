package block

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/tests"
	"golang.org/x/crypto/sha3"
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

	_, blkbytes, e := db.GetBlockBytesByHeight(rootblk.GetHeight(), true, true, 0)
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

		_, blkbytes, e := db.GetBlockBytesByHeight(rootblk.GetHeight(), true, true, 0)
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

func Test_mkrl_hash(t *testing.T) {

	hashs := make([][]byte, 5)
	hashs[0] = make([]byte, 32)
	hashs[1] = make([]byte, 32)
	hashs[2] = make([]byte, 32)
	hashs[3] = make([]byte, 32)
	hashs[4] = make([]byte, 32)

	random := make([]byte, 32)
	rand.Read(random)
	copy(hashs[0], random)
	rand.Read(random)
	copy(hashs[1], random)
	rand.Read(random)
	copy(hashs[2], random)
	rand.Read(random)
	copy(hashs[3], random)
	rand.Read(random)
	copy(hashs[4], random)

	var root []byte
	for {
		if len(hashs) == 1 {
			root = hashs[0]
			break
		}
		hashs = hashMerge(hashs) // 两两归并
	}

	fmt.Println("====================")

	fmt.Println(hex.EncodeToString(root))

}

func hashMerge(hashs [][]byte) [][]byte {
	fmt.Println("---------------")
	length := len(hashs)
	mgsize := length / 2
	if length%2 == 1 {
		mgsize = (length + 1) / 2
	}
	var mergehashs = make([][]byte, mgsize)
	for m := 0; m < length; m += 2 {
		var b1 bytes.Buffer
		b1.Write(hashs[m])
		h2 := hashs[m]
		if m+1 < length {
			h2 = hashs[m+1]
		}
		b1.Write(h2)
		fmt.Println(hex.EncodeToString(hashs[m]), "\n"+hex.EncodeToString(h2)+"\n")
		digest := sha3.Sum256(b1.Bytes())
		mergehashs[m/2] = digest[:]
	}
	return mergehashs
}
