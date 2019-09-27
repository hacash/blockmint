package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/types/block"
	"strconv"
	"strings"
)

// 通过 高度 或 hx 获取区块简介
func getBlockIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	var isgettxhxs = false // 是否获取区块交易hash列表
	if _, ok0 := params["gettrshxs"]; ok0 {
		isgettxhxs = true
	}
	blkid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}
	blkdb := store.GetGlobalInstanceBlocksDataStore()
	var blockhx = []byte{}
	var blockbytes = []byte{}
	if blkhei, err := strconv.ParseUint(blkid, 10, 0); err == nil {
		bhx, bodys, e := blkdb.GetBlockBytesByHeight(blkhei, true, isgettxhxs, 0)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		blockhx = bhx
		blockbytes = bodys
	} else if bhx, e := hex.DecodeString(blkid); e == nil && len(bhx) == 32 {
		blockhx = bhx
		var e error
		if isgettxhxs {
			blockbytes, e = blkdb.ReadBlockBytes(bhx)
		} else {
			blockbytes, e = blkdb.ReadHeadBytes(bhx)
		}
		if e != nil {
			result["err"] = e.Error()
			return result
		}
	} else {
		result["err"] = "block id <" + blkid + "> not find."
		result["ret"] = "1"
		return result
	}
	// 解析区块
	var tarblock block.Block
	var err error
	if isgettxhxs {
		tarblock, _, err = blocks.ParseBlock(blockbytes, 0)
	} else {
		tarblock, _, err = blocks.ParseBlockHead(blockbytes, 0)
	}
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	// 区块返回数据
	result["jsondata"] = fmt.Sprintf(
		`{"hash":"%s","height":%d,"prevhash":"%s","mrklroot":"%s","timestamp":%d,"txcount":%d,"reward":"%s"`,
		hex.EncodeToString(blockhx),
		tarblock.GetHeight(),
		hex.EncodeToString(tarblock.GetPrevHash()),
		hex.EncodeToString(tarblock.GetMrklRoot()),
		tarblock.GetTimestamp(),
		tarblock.GetTransactionCount(),
		coin.BlockCoinBaseReward(tarblock.GetHeight()).ToFinString(), // 奖励数量
	)
	// 区块hx列表
	if isgettxhxs {
		var blktxhxsary []string
		var blktxhxsstr = ""
		var rwdaddr fields.Address // 奖励地址
		for i, trs := range tarblock.GetTransactions() {
			if i == 0 {
				rwdaddr = fields.Address(trs.GetAddress())
				blktxhxsary = append(blktxhxsary, "[coinbase]")
			} else {
				blktxhxsary = append(blktxhxsary, hex.EncodeToString(trs.HashNoFee()))
			}
		}
		blktxhxsstr = strings.Join(blktxhxsary, `","`)
		if len(blktxhxsstr) > 0 {
			blktxhxsstr = `"` + blktxhxsstr + `"`
		}
		result["jsondata"] += fmt.Sprintf(
			`,"nonce":%d,"difficulty":%d,"rwdaddr":"%s","trshxs":[%s]`,
			tarblock.GetNonce(),
			tarblock.GetDifficulty(),
			rwdaddr.ToReadable(),
			blktxhxsstr,
		)
	}
	// 收尾并返回
	result["jsondata"] += "}"
	return result
}

// 获取最新区块高度
func getLastBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)
	minerstate := miner.GetGlobalInstanceHacashMiner().State
	result["jsondata"] = fmt.Sprintf(
		`{"height":%d,"txs":%d,"timestamp":%d}`,
		minerstate.CurrentHeight(),
		minerstate.GetBlockHead().GetTransactionCount()-1,
		minerstate.GetBlockHead().GetTimestamp(),
	)
	return result
}

// 获取区块摘要信息
func getBlockAbstractList(params map[string]string) map[string]string {
	result := make(map[string]string)
	start, ok1 := params["start_height"]
	end, ok2 := params["end_height"]
	if !ok1 || !ok2 {
		result["err"] = "start_height or end_height must"
		return result
	}
	start_hei, e1 := strconv.ParseUint(start, 10, 0)
	end_hei, e2 := strconv.ParseUint(end, 10, 0)
	if e1 != nil || e2 != nil {
		result["err"] = "start_height or end_height param error"
		return result
	}
	if end_hei-start_hei+1 > 100 {
		result["err"] = "start_height - end_height cannot more than 100"
		return result
	}
	// 查询区块信息
	db := store.GetGlobalInstanceBlocksDataStore()
	coinbase_head_len := uint32(1 + 21 + 3 + 15) // 16 drop tail
	var jsondata = make([]string, 0, end_hei-start_hei+1)
	for i := end_hei; i >= start_hei; i-- {
		blkhash, blkbytes, e := db.GetBlockBytesByHeight(i, true, true, 10+coinbase_head_len)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		blkhead, _, e2 := blocks.ParseExcludeTransactions(blkbytes, 0)
		if e2 != nil {
			result["err"] = e2.Error()
			return result
		}
		// 解析矿工信息
		var coinbase transactions.Transaction_0_Coinbase
		coinbase.ParseHead(blkbytes, uint32(len(blkbytes))-coinbase_head_len+1)
		coinbase.Message = fields.TrimString16([]byte(coinbase.Message)[0:12])
		// 返回
		jsondata = append(jsondata, fmt.Sprintf(
			`{"hash":"%s","txs":%d,"time":%d,"height":%d,"nonce":%d,"bits":%d,"rewards":{"amount":"%s","address":"%s","message":"%s"}}`,
			hex.EncodeToString(blkhash),
			blkhead.GetTransactionCount()-1,
			blkhead.GetTimestamp(),
			blkhead.GetHeight(),
			blkhead.GetNonce(),
			blkhead.GetDifficulty(),
			coinbase.Reward.ToFinString(),
			coinbase.Address.ToReadable(),
			strings.Replace(string(bytes.Trim([]byte(coinbase.Message), string([]byte{0}))), `"`, ``, -1),
		))
		//addrbytes = bytes.Trim(addrbytes, string([]byte{0}))
		//fmt.Println([]byte(coinbase.Message))
	}
	// 返回
	result["jsondata"] = `{"datas":[` + strings.Join(jsondata, ",") + `]}`
	return result
}
