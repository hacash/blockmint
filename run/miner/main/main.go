package main

import (
	"fmt"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/service/rpc"
	"os"
	"os/signal"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// config
	config.LoadConfigFile()

	// http 接口
	go rpc.RunHttpRpcService()

	var miner = miner.GetGlobalInstanceHacashMiner()
	go miner.Start() // 准备挖矿
	if config.Config.Miner.Forcestart == "true" {
		miner.CanStart() // 开始挖矿
	}

	var ptcmng = p2p2.GetGlobalInstanceProtocolManager()
	go ptcmng.Start(0)

	var p2p = p2p2.GetGlobalInstanceP2PServer()
	go p2p.Start() // 加入p2p网络

	s := <-c
	fmt.Println("Got signal:", s)

}
