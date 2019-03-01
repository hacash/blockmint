package main

import (
	"github.com/hacash/blockmint/config"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/service/rpc"
)

func main() {

	// config
	config.LoadConfigFile()

	// http 接口
	go rpc.RunHttpRpcService()
	go func() {
		// var miner = miner.GetGlobalInstanceHacashMiner()
		// miner.Start() // 开始挖矿
	}()
	go func() {
		var p2p = p2p2.GetGlobalInstanceP2PServer()
		p2p.Start() // 加入p2p网络
	}()

	select {}

}
