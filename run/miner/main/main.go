package main

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/chain/state"
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
	//Test_coinbaseAddress(16231)
	//Test_opencl()

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

/*
var kernelSource = `
__kernel void square(
   __global float* input,
   __global float* output,
   const unsigned int count)
{
   int i = get_global_id(0);
   if(i < count)
       output[i] = input[i] * input[i];
}
`

// 测试opencl编程
func Test_opencl() {

	var data [1024]float32
	for i := 0; i < len(data); i++ {
		data[i] = rand.Float32()
	}

	platforms, err := cl.GetPlatforms()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(platforms) == 0 {
		fmt.Println("platforms.length = 0")
		return
	}
	for _, plat := range platforms {
		fmt.Println(plat.Name())
	}
	platform := platforms[0]
	devices, err := platform.GetDevices(cl.DeviceTypeAll)
	for i, d := range devices {
		fmt.Println("Device %d (%s): %s", i, d.Type(), d.Name())
	}
	device := devices[0]
	context, err1 := cl.CreateContext([]*cl.Device{device})
	if err1 != nil {
		fmt.Println("CreateContext failed: %+v", err1)
		return
	}
	queue, err2 := context.CreateCommandQueue(device, 0)
	if err2 != nil {
		fmt.Println("CreateCommandQueue failed: %+v", err2)
		return
	}
	program, err3 := context.CreateProgramWithSource([]string{kernelSource})
	if err3 != nil {
		fmt.Println("CreateProgramWithSource failed: %+v", err3)
		return
	}
	if err4 := program.BuildProgram(nil, ""); err4 != nil {
		fmt.Println("BuildProgram failed: %+v", err4)
		return
	}
	kernel, err5 := program.CreateKernel("square")
	if err5 != nil {
		fmt.Println("CreateKernel failed: %+v", err5)
		return
	}
	for i := 0; i < 3; i++ {
		name, err := kernel.ArgName(i)
		if err == cl.ErrUnsupported {
			break
		} else if err != nil {
			fmt.Println("GetKernelArgInfo for name failed: %+v", err)
			break
		} else {
			fmt.Println("Kernel arg %d: %s", i, name)
		}
	}
	input, err6 := context.CreateEmptyBuffer(cl.MemReadOnly, 4*len(data))
	if err6 != nil {
		fmt.Println("CreateBuffer failed for input: %+v", err6)
		return
	}
	output, err7 := context.CreateEmptyBuffer(cl.MemReadOnly, 4*len(data))
	if err7 != nil {
		fmt.Println("CreateBuffer failed for output: %+v", err7)
		return
	}
	if _, err8 := queue.EnqueueWriteBufferFloat32(input, true, 0, data[:], nil); err8 != nil {
		fmt.Println("EnqueueWriteBufferFloat32 failed: %+v", err8)
		return
	}
	if err9 := kernel.SetArgs(input, output, uint32(len(data))); err9 != nil {
		fmt.Println("SetKernelArgs failed: %+v", err9)
		return
	}

	local, err10 := kernel.WorkGroupSize(device)
	if err10 != nil {
		fmt.Println("WorkGroupSize failed: %+v", err10)
		return
	}
	fmt.Println("Work group size: %d", local)
	size, _ := kernel.PreferredWorkGroupSizeMultiple(nil)
	fmt.Println("Preferred Work Group Size Multiple: %d", size)

	global := len(data)
	d := len(data) % local
	if d != 0 {
		global += local - d
	}
	if _, err := queue.EnqueueNDRangeKernel(kernel, nil, []int{global}, []int{local}, nil); err != nil {
		fmt.Println("EnqueueNDRangeKernel failed: %+v", err)
		return
	}

	if err := queue.Finish(); err != nil {
		fmt.Println("Finish failed: %+v", err)
		return
	}

	results := make([]float32, len(data))
	if _, err := queue.EnqueueReadBufferFloat32(output, true, 0, results, nil); err != nil {
		fmt.Println("EnqueueReadBufferFloat32 failed: %+v", err)
		return
	}
	fmt.Println(results)

	correct := 0
	for i, v := range data {
		if results[i] == v*v {
			correct++
		}
	}

	if correct != len(data) {
		fmt.Println("%d/%d correct values", correct, len(data))
	}

	fmt.Println("==========================")

}
*/

//
// 测试打印区块奖励地址
func Test_coinbaseAddress(height uint64) {

	_, blkbts, _ := store.GetGlobalInstanceBlocksDataStore().GetBlockBytesByHeight(height, true, true, 0)
	blk, _, _ := blocks.ParseBlock(blkbts, 0)

	trs := blk.GetTransactions()
	if coinbase, ok := trs[0].(*transactions.Transaction_0_Coinbase); ok {

		addr := base58check.Encode(coinbase.Address)
		fmt.Println(addr, coinbase.Reward.ToFinString())

	}

}

// 测试打印区块奖励地址
func Test_database_store(height uint64) {

	blkbts, _ := hex.DecodeString("010000003f37005c90a5b80000000d0d0af1c87d65c581310bd7ae803b23c69754be16df02a7b156c03c87aadd0ada0615668c7bf3658efeab80ef2a6be1e884a2844d52afdb88fa82f5c6000000010070db79e48fffa400000000ff89de02003bea1b64e8d5659d314c078ad37551f801012020202020202020202020202020202000")
	blk, _, _ := blocks.ParseBlock(blkbts, 0)
	// 保存
	sss := state.GetGlobalInstanceChainState()
	ssstemp := state.NewTempChainState(sss)
	blk.ChangeChainState(ssstemp)
	sss.TraversalCopy(ssstemp)

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
		_, blkbts, _ := db.GetBlockBytesByHeight(i, true, true, 0)
		block, _, _ := blocks.ParseBlock(blkbts, 0)
		coinbase, _ := block.GetTransactions()[0].(*transactions.Transaction_0_Coinbase)
		addr := base58check.Encode(coinbase.Address)
		if _, ok := rewards[addr]; ok {
			rewards[addr] += 1
		} else {
			rewards[addr] = 1
		}
	}
	var state = state.GetGlobalInstanceChainState()
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
