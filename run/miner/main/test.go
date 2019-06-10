package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/miner"
	"os"
	"strings"
)


// 测试打印全部的钻石关系
func Test_allAddressDiamonds() {

	diamondDB := db.GetGlobalInstanceDiamondDB()

	diamondsfile, fe := os.OpenFile("./diamonds.idx", os.O_RDWR, 0777)
	if fe != nil {
		panic(fe)
	}


	diastat, _ := diamondsfile.Stat()
	signum := diastat.Size() / 6
	diastr := make([]byte, 6)
	basebts := []byte("WTYUIAHXVMEKBSZN")
	asddressDiamonds := make(map[string][]string)
	for i:=int64(0); i<signum; i ++ {
		diamondsfile.ReadAt(diastr, i * 6)
		if bytes.IndexByte(basebts, diastr[0]) == -1 {
			continue
		}
		// 查询钻石
		addr, _ := diamondDB.Read(diastr)
		if addr != nil {
			key := addr.ToReadable()
			if _, ok := asddressDiamonds[key]; !ok {
				asddressDiamonds[key] = make([]string, 0, 16)
			}
			asddressDiamonds[key] = append(asddressDiamonds[key], string(diastr))
		}
	}
	// 打印全部钻石所属
	total_num := 0
	for k, dias := range asddressDiamonds {
		fmt.Print("\n"+k+": ")
		for _, v := range dias {
			fmt.Print(v+",")
			total_num++
		}
	}
	fmt.Println("\n TOTAL: ", total_num)


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



func Test_address_balance() {

	// 批量查询地址余额
	addrstr := `19nZSSpt5Qh5wrFY7EiXWLamnTULDcJnLa,1Bopq1R7asXzCg1caSX66D67WVGW3z1Z92,16VYWXKfE4MxRKFjgLWgBywrzH2LUipw2,1GnszPejCXFDxQcfD6KK5MkDqfB71DPtYY,12Hp4mTvhh6RytKD9G4oRmrCQ6voUc6gNN,16o19uBcYByy2EJBo2iMnpXNb4vnUMZ7BN,1AET4VPWDEGujSbKsKkcs1mVW9jbf971hj,1NM1oHX3NkQr3xDRPdoDqSEQZopMdEmgtP,1D6bMGntxDYnExYUNtxED3ffvA2ne79fRW,14tc1GR54ko89xbwzaftMrGQ75KUqJmuT8,15ekGGRkGHpyxxJYQP4KXLmnqCkiyUmdKG,1fWgwWs3K5PPzAajvkFRRxTb113BYmLxe,1M9H29pGrGXjNwzjrWJNYU6NFyFTStymF7,179i51n1Tv5DqJCbfrvpNpWK3WeCoMR313,1KA83xNHkajrMnjAZj2ANdQnGU2GnLmP6T,1Kjy1c14h2ooU5CJS5AdNpRhZSdTKDvP7k,13TYqp95XQws9JZvRQBau16gpP5sLhWLfF,16RfEb8YjZdRYBPUaRiYuQbcUByFMcU599,14bjJWuNMRZC2zNnEPL8Aw43hDXDFgxgV9,1GQH4GDARuvB6nApHJeWVh2fRBxKYYjdx5,1BNrAjZfckcMzDJANc3TLqUktekTnPdn7k,1FiU7NjFX2VhkCT213q3yrQMiEms154G2M,138wA7SRNejWFhwqCN6f4bdNvviz46E2cQ,1FQczPeo96LgzzCMaDAFSvBLKfrzKqXWAi,1aVcodF5SC9uoWtCQ3bz4j7dtTV2NNw1d,1BjbnHwhV7VgL4kM3EsEHjyjwF5MGRNS3f,14DQn8EdHpo53dhvbTPMn8VgYVmKvNiXG4,14pNWhTLLPwWsdvrwFjpNGWR9X8WSPe5x3`

	addrs := strings.Split(addrstr, ",")
	blsdb := db.GetGlobalInstanceBalanceDB()

	fmt.Println("-------------------------------------------")
	for _, v := range addrs {
		v = strings.Trim(v, " ")
		address, _ := fields.CheckReadableAddress(v)
		if address == nil {
			continue
		}
		balance, _ := blsdb.Read(*address)
		if balance != nil {
			fmt.Println(v, balance.Amount.ToFinString())
		}
	}
	fmt.Println("-------------------------------------------")

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
