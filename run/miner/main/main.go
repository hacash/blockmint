package main

import (
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/service/rpc"
)

func main() {

	chwait := make(chan int, 1)

	// http 接口
	go rpc.RunHttpRpcService()
	go func() {
		var miner = miner.GetGlobalInstanceHacashMiner()
		miner.Start() // 开始挖矿
	}()

	<-chwait

}
