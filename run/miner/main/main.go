package main

import "github.com/hacash/blockmint/service/rpc"

func main() {

	chwait := make(chan int, 1)

	// http 接口
	go rpc.RunHttpRpcService()

	<-chwait

}
