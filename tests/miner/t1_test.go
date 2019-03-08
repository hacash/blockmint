package miner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	base58check2 "github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/sys/log"
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

	var miner = miner.NewHacashMiner(log.New())
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

	//var miner = miner.NewHacashMiner()
	//miner.Start()

	fmt.Println(hex.DecodeString("57cef097f9a7cc0c45bcac6325b5b6e58199c8197763734cac6664e8d2b8e63e"))

	blk := coin.GetGenesisBlock()

	fmt.Println(blk.Hash())

	address, _ := base58check2.Decode("13337681iZShfzYDkNBmjnaoAfchsrFQpD")
	fmt.Println(address)

	add := []byte{0, 22, 82, 215, 51, 169, 135, 248, 255, 189, 82, 120, 227, 122, 205, 117, 64, 162, 253, 1, 38}

	if bytes.Compare(add, address) != 0 {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}

}
