package rpc

import (
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/miner"
	"strconv"
	"strings"
)

// 扫描区块 获取所有转账信息
func getAllTransferLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)
	block_height_str, ok1 := params["block_height"]
	if !ok1 {
		result["err"] = "param block_height must."
		return result
	}
	block_height, err2 := strconv.ParseUint(block_height_str, 10, 0)
	if err2 != nil {
		result["err"] = "param block_height format error."
		return result
	}
	// 判断区块高度
	if block_height <= 0 || block_height > miner.GetGlobalInstanceHacashMiner().State.CurrentHeight() {
		result["err"] = "block height not find."
		result["ret"] = "1" // 返回错误码
		return result
	}

	// 查询区块
	db := store.GetGlobalInstanceBlocksDataStore()
	_, targetblockdata, err3 := db.GetBlockBytesByHeight(block_height, true, true, 0)
	if err3 != nil {
		result["err"] = "read block data error."
		return result
	}
	tarblock, _, err4 := blocks.ParseBlock(targetblockdata, 0)
	if err4 != nil {
		result["err"] = "block data parse error."
		return result
	}

	// 开始扫描区块
	allTransferLogs := make([]string, 0, 4)
	transactions := tarblock.GetTransactions()
	for _, v := range transactions {
		if 1 == v.Type() { // coinbase
			continue
		}
		from := fields.Address(v.GetAddress())
		for _, act := range v.GetActions() {
			if 1 == act.Kind() { // 类型为普通转账
				act_k1 := act.(*actions.Action_1_SimpleTransfer)
				allTransferLogs = append(allTransferLogs,
					from.ToReadable()+","+
						act_k1.Address.ToReadable()+","+
						act_k1.Amount.ToFinString())
			}
		}
	}

	datasstr := strings.Join(allTransferLogs, "\",\"")
	if len(datasstr) > 0 {
		datasstr = "\"" + datasstr + "\""
	}

	// 返回
	result["jsondata"] = `{"timestamp":` + strconv.FormatUint(tarblock.GetTimestamp(), 10) + `,"datas":[` + datasstr + `]}`
	return result
}
