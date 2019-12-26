package rpc

import (
	"encoding/hex"
	"fmt"
	actions2 "github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"strings"
)

// 通过 hx 获取交易简介
func getTransactionIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	trsid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}
	var trshx []byte
	if txhx, e := hex.DecodeString(trsid); e == nil && len(txhx) == 32 {
		trshx = txhx
	} else {
		result["err"] = "transaction hash error."
		return result
	}
	// 查询交易
	trsdb := store.GetGlobalInstanceBlocksDataStore()
	trsres, err := trsdb.ReadTransaction(trshx, true, true)
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	if trsres == nil {
		result["err"] = "transaction not fond."
		return result
	}
	// 解析 actions
	var actions = trsres.Transaction.GetActions()
	var actions_ary []string
	var actions_strings = ""
	for _, act := range actions {
		var kind = act.Kind()
		actstr := fmt.Sprintf(`{"k":%d`, kind)
		if kind == 1 {
			acc := act.(*actions2.Action_1_SimpleTransfer)
			actstr += fmt.Sprintf(`,"to":"%s","amount":"%s"`,
				acc.Address.ToReadable(),
				acc.Amount.ToFinString(),
			)
		}
		actstr += "}"
		actions_ary = append(actions_ary, actstr)
	}
	actions_strings = strings.Join(actions_ary, ",")
	// 交易返回数据
	txaddr := fields.Address(trsres.Transaction.GetAddress())
	var txfee fields.Amount
	txfee.Parse(trsres.Transaction.GetFee(), 0)
	result["jsondata"] = fmt.Sprintf(
		`{"block":{"height":%d,"timestamp":%d},"type":%d,"address":"%s","fee":"%s","timestamp":%d,"actioncount":%d,"actions":[%s]`,
		trsres.BlockHead.GetHeight(),
		trsres.BlockHead.GetTimestamp(),
		trsres.Transaction.Type(),
		txaddr.ToReadable(), // 主地址
		txfee.ToFinString(),
		trsres.Transaction.GetTimestamp(),
		len(actions),
		actions_strings,
	)

	// 收尾并返回
	result["jsondata"] += "}"
	return result
}
