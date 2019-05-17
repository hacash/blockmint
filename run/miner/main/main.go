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
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner"
	p2p2 "github.com/hacash/blockmint/p2p"
	"github.com/hacash/blockmint/run/diamond"
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
	/*
		nnn := make([]byte, 4)
		binary.BigEndian.PutUint32(nnn, blk.GetNonce())
		fmt.Println(nnn )
		fmt.Println(blk.HashFresh())
		fmt.Println( hex.EncodeToString(blk.HashFresh()) )
		// nonce: [0 112 219 121]
		// hash: [0 0 0 10 198 94 211 152 131 143 206 7 61 245 177 81 50 218 67 111 126 41 147 53 63 211 102 43 248 178 207 145]
		// hash: 0000000ac65ed398838fce073df5b15132da436f7e2993353fd3662bf8b2cf91
	*/
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



func Test_others()  {


	//amt1, _ := fields.NewAmountFromFinString("HCX1:248")
	//amt2 := fields.NewAmountSmall(1,248)
	//actions.DoAppendCompoundInterest1Of10000By2500Height(amt1, amt2, 142)

	//fmt.Println( x16rs.CheckDiamondDifficulty(1, []byte{0,0,255,255}) )
	//fmt.Println( x16rs.CheckDiamondDifficulty(2048*256*2 + 2048+1, []byte{0,0,254,255}) )

	//sss := state.GetGlobalInstanceChainState()
	//cid, _ := hex.DecodeString("f89d8e27bf6c5c8c6b236221210593ee")
	//res := sss.Channel(fields.Bytes16(cid))
	//fmt.Println(res)

	//nonce, _ := hex.DecodeString("010000005689050d") // 010000005689050d
	//prevHash, _ := hex.DecodeString("000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e6")
	//address, _ := fields.CheckReadableAddress("1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9") // 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9
	//r1, r2 := x16rs.Diamond(uint32(1), prevHash, nonce, *address)
	//fmt.Println(hex.EncodeToString(r1),  r2)

}