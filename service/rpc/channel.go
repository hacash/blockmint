package rpc

import (
	"encoding/hex"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func getChannel(params map[string]string) map[string]string {
	result := make(map[string]string)
	idstr, ok1 := params["ids"]
	if !ok1 {
		result["err"] = "params ids must."
		return result
	}
	idlist := strings.Split(idstr, ",")
	if len(idlist) == 0 {
		result["err"] = "params ids must."
		return result
	}
	total_amount := fields.NewEmptyAmount()
	for i := 0; i < len(idlist); i++ {
		idstr := idlist[i]

		if len(idstr) != 32 {
			result["fail"] = "id format error."
			return result
		}
		chanid, e0 := hex.DecodeString(idstr)
		if e0 != nil {
			result["fail"] = "id format error."
			return result
		}

		dmdb := db.GetGlobalInstanceChannelDB()
		store, e1 := dmdb.Read(fields.Bytes16(chanid))
		if e1 != nil || store == nil {
			result["fail"] = "not find."
			return result
		}
		totalamt, _ := store.LeftAmount.Add(&store.RightAmount)
		if len(idlist) == 1 {
			// 只有一条数据则返回详情
			result["is_closed"] = strconv.Itoa(int(store.IsClosed))
			result["belong_height"] = strconv.FormatUint(uint64(store.BelongHeight), 10)
			result["left_address"] = store.LeftAddress.ToReadable()
			result["left_amount"] = store.LeftAmount.ToFinString()
			result["right_address"] = store.RightAddress.ToReadable()
			result["right_amount"] = store.RightAmount.ToFinString()
			result["total_amount"] = totalamt.ToFinString()
			return result
		} else {
			// 否则返回加总统计
			total_amount, _ = total_amount.Add(totalamt)
		}
	}
	// 返回总计
	result["total"] = strconv.Itoa(len(idlist))
	result["total_amount"] = total_amount.ToFinString()
	return result
}
