package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	state2 "github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/service/rpc"
	"os"
	"os/signal"
	"time"
)


/**
 * go build -o miner_node_hacash run/miner/main/main.go && ./miner_node_hacash
 */


func main() {

	// config
	config.LoadConfigFile()

	//Test_coinbaseAmt()
	//Test_coinbaseAddress(16183)

	StartHacash()

}

// 启动
func StartHacash() {

	if config.Config.Miner.Backtoheight > 0 {
		tarhei := config.Config.Miner.Backtoheight
		// 区块状态倒退
		fmt.Println("go back the block chain data state to the specified height ", tarhei)
		var miner = miner.GetGlobalInstanceHacashMiner()
		_, err := miner.BackTheWorldToHeight(tarhei)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("ok, back to height", tarhei, "now.")
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
			fmt.Println("HacashMiner start mining in force on start...")
			t := time.NewTimer(5 * time.Second)
			<-t.C
			fmt.Println("start mining...")
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


//


// 测试打印区块奖励地址
func Test_coinbaseAddress(height uint64) {

	blkbts, _ := hex.DecodeString("010000003f37005c90a5b80000000d0d0af1c87d65c581310bd7ae803b23c69754be16df02a7b156c03c87aadd0ada0615668c7bf3658efeab80ef2a6be1e884a2844d52afdb88fa82f5c6000000010070db79e48fffa400000000ff89de02003bea1b64e8d5659d314c078ad37551f801012020202020202020202020202020202000")
	blk, _, _ := blocks.ParseBlock(blkbts, 0)
	// 保存
	sss := state2.GetGlobalInstanceChainState()
	ssstemp := state2.NewTempChainState( sss )
	blk.ChangeChainState( ssstemp )
	sss.TraversalCopy( ssstemp )

	trs := blk.GetTransactions()
	if coinbase, ok := trs[0].(*transactions.Transaction_0_Coinbase); ok {
		fmt.Println("hash", hex.EncodeToString(blk.HashFresh()), "prev", hex.EncodeToString(blk.GetPrevHash()))
		addr := base58check.Encode(coinbase.Address)
		fmt.Println(addr, coinbase.Reward.ToFinString())
		amtread := sss.Balance(coinbase.Address)
		fmt.Println("111111111111111111111111111111111111111")
		fmt.Println(amtread.ToFinString())
		fmt.Println("222222222222222222222222222222222222222")
		amtread2 := ssstemp.Balance(coinbase.Address)
		fmt.Println(amtread2.ToFinString())




		blcdb := db.GetGlobalInstanceBalanceDB()
		finditem, e1 := blcdb.Read(coinbase.Address)
		if e1 != nil {
			fmt.Println(e1)
		}
		if finditem != nil {
			fmt.Println("amount", finditem.Amount.ToFinString())
		}




	}





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
