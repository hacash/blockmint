package main

import (
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	state2 "github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/service/rpc"
	"os"
	"os/signal"
	"time"
)

func main() {

	// config
	config.LoadConfigFile()

	//Test_coinbaseAmt()
	//return

	StartHacash()

}

// 启动
func StartHacash() {

	if config.Config.Miner.Backtoheight > 0 {
		tarhei := config.Config.Miner.Backtoheight
		// 区块状态倒退
		fmt.Println("Back the block chain data state to the specified height ", tarhei)
		var miner = miner.GetGlobalInstanceHacashMiner()
		_, err := miner.BackTheWorldToHeight(tarhei)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Ok, back to height", tarhei, "now.")
		}
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// http 接口
	go rpc.RunHttpRpcService()

	var miner = miner.GetGlobalInstanceHacashMiner()
	go miner.Start()
	if config.Config.Miner.Forcestart == "true" {
		go func() {
			fmt.Println("HacashMiner start mining in force on start ...")
			t := time.NewTimer(5 * time.Second)
			<-t.C
			miner.StartMining() // 开始挖矿
		}()
	}

	var ptcmng = p2p2.GetGlobalInstanceProtocolManager()
	go ptcmng.Start(0)

	var p2p = p2p2.GetGlobalInstanceP2PServer()
	go p2p.Start() // 加入p2p网络

	s := <-c
	fmt.Println("Got signal:", s)

}

// 测试打印余额
func Test_coinbaseAmt() {

	//////// COUNT
	var db = store.GetGlobalInstanceBlocksDataStore()
	curheight := miner.GetGlobalInstanceHacashMiner().State.CurrentHeight()
	rewards := make(map[string]int)
	for i := uint64(1); i <= curheight; i++ {
		blkbts, _ := db.GetBlockBytesByHeight(i, true, true)
		block, _, _ := blocks.ParseBlock(blkbts, 0)
		coinbase, _ := block.GetTransactions()[0].(*transactions.Transaction_0_Coinbase)
		addr := base58check.Encode(coinbase.Address)
		if _, ok := rewards[addr]; ok {
			rewards[addr] += 1
		} else {
			rewards[addr] = 1
		}
	}
	var state = state2.GetGlobalInstanceChainState()
	total := 0
	totalAmt := fields.NewEmptyAmount()
	for k, v := range rewards {
		address, _ := base58check.Decode(k)
		amt := state.Balance(address)
		totalAmt, _ = totalAmt.Add(&amt)
		fmt.Println(k, v, amt.ToFinString()) //, amt.Unit, amt.Dist, amt.Numeral)
		total += v
	}
	fmt.Println("total", total, totalAmt.ToFinString())

	//////// COUNT END
}
