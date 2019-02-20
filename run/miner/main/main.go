package main

import "github.com/hacash/blockmint/service/rpc"

func main() {

	chwait := make(chan int, 1)

	go rpc.RunHttpRpcService()

	<-chwait

}
