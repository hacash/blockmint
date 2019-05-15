package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/miner"
	"strconv"
	"strings"
)

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
