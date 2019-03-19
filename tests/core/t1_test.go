package miner

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/core/coin"
	"testing"
	"time"
)

func Test_t1(t *testing.T) {

	for i := uint64(0); ; i++ {
		h := i*10000*10 + 99999
		rw := coin.BlockCoinBaseReward(h)
		fmt.Printf("%d: %s \n", h, rw.ToFinString())
		time.Sleep(time.Duration(33) * time.Millisecond)
	}

}

func Test_t2(t *testing.T){

	dist, _ := hex.DecodeString("000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e6")

	for i:=0; i<10000*1000; i++ {
		blk := coin.GetGenesisBlock()
		hash :=  blk.HashFresh()
		if bytes.Compare( hash, dist ) != 0 {
			fmt.Println( hex.EncodeToString( hash ) )
		}

	}


}