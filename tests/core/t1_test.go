package miner

import (
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
