package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/chain/state"
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
 * go build -o mmmhcx3 run/miner/main/main.go && ./mmmhcx3 run/miner/main/localtestcnf/3.yml
 * go build -o mmmhcx4 github.com/hacash/blockmint/run/miner/main/ && ./mmmhcx4 run/miner/main/localtestcnf/4.yml
 */

func main() {

	fmt.Println("net genesis block is ", hex.EncodeToString(coin.GetGenesisBlock().HashFresh()))

	// config
	config.LoadConfigFile()

	//Test_1BitcoinMoveToHacashNeverBack()
	//Test_coinbaseAmt()
	//Test_coinbaseAddress(16231)
	//Test_opencl()
	//Test_address_balance()
	//Test_allAddressDiamonds()
	//Test_allAddressAmount()
	//Test_allChannelAmount()

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

	var minerobj = miner.GetGlobalInstanceHacashMiner()
	go minerobj.Start()
	if config.Config.Miner.Forcestart == "true" {
		go func() {
			fmt.Println("HacashMiner start mining in force on start ...")
			t := time.NewTimer(5 * time.Second)
			<-t.C
			minerobj.StartMining() // 开始挖矿
		}()
	}

	// 设置矿工状态
	state.GetGlobalInstanceChainState().SetMiner(minerobj)

	var ptcmng = p2p2.GetGlobalInstanceProtocolManager()
	go ptcmng.Start(0)

	var p2p = p2p2.GetGlobalInstanceP2PServer()
	go p2p.Start() // 加入p2p网络

	// 如果挖掘钻石
	if len(config.Config.DiamondMiner.Feepassword) > 6 {
		dm := diamond.NewDiamondMiner()
		fmt.Println("❂ start diamond mining...")
		go dm.Start(minerobj) // 开始挖掘
	}

	if len(config.Config.MiningPool.StatisticsDir) > 0 {
		if len(config.Config.MiningPool.PayPassword) < 6 {
			panic("Config.MiningPool.PayPassword length must more than 6")
		}
		if config.Config.MiningPool.PayFeeRatio < 0 || config.Config.MiningPool.PayFeeRatio >= 1 {
			panic("Config.MiningPool.PayFeeRatio value format error")
		}
		// 矿池启动
		miningpool := miner.GetGlobalInstanceMiningPool()
		go miningpool.Start() // 启动
	}

	s := <-c
	fmt.Println("Got signal:", s)

}
