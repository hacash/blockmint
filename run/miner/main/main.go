package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/run/miner/diamond"
	"github.com/hacash/blockmint/service/rpc"
	"os"
	"os/signal"
	"time"
)

/**
 * go build -o miner_node_hacash run/miner/main/main.go && ./miner_node_hacash
 */

func main() {

	fmt.Println("net genesis block is ", hex.EncodeToString(coin.GetGenesisBlock().HashFresh()))

	// config
	config.LoadConfigFile()

	//Test_coinbaseAmt()
	//Test_coinbaseAddress(16231)
	//Test_opencl()
	//Test_address_balance()
	//Test_allAddressDiamonds()


	

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

	// 如果挖掘钻石
	if len(config.Config.DiamondMiner.Feepassword) > 6 {
		dm := diamond.NewDiamondMiner()
		fmt.Println("❂ start diamond mining...")
		go dm.Start(miner) // 开始挖掘
	}

	s := <-c
	fmt.Println("Got signal:", s)

}

