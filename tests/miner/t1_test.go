package miner

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner"
	"testing"
)

func Test_t2(t *testing.T) {

	for i := 0; i < 100; i++ {
		genis := coin.GetGenesisBlock()
		hash := genis.Hash()
		fmt.Println(hex.EncodeToString(hash))
	}

}

func Test_t1(t *testing.T) {

	genis := coin.GetGenesisBlock()
	dttypes, _ := genis.Serialize()

	hash := genis.Hash()
	fmt.Println(hash)
	fmt.Println(hex.EncodeToString(hash))
	fmt.Println(dttypes)
	fmt.Println(hex.EncodeToString(dttypes))
	fmt.Println(len(dttypes))

	var miner = miner.NewHacashMiner()
	nextblk, _ := miner.CreateBlock()
	nexttypes, _ := nextblk.Serialize()
	hash2 := nextblk.Hash()
	fmt.Println(hash2)
	fmt.Println(hex.EncodeToString(hash2))
	fmt.Println(nexttypes)
	fmt.Println(hex.EncodeToString(nexttypes))
	fmt.Println(len(nexttypes))

}

func Test_t3(t *testing.T) {

	var miner = miner.NewHacashMiner()
	miner.Start()

}
