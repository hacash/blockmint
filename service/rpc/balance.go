package rpc

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"strings"
)

//////////////////////////////////////////////////////////////

func getBalance(params map[string]string) map[string]string {
	result := make(map[string]string)
	addrstr, ok1 := params["address"]
	if !ok1 {
		result["err"] = "address must"
		return result
	}
	blcdb := db.GetGlobalInstanceBalanceDB()
	addrs := strings.Split(addrstr, ",")
	amtstrings := ""
	totalamt := fields.NewEmptyAmount()
	for k, addr := range addrs {
		if k > 20 {
			break // 最多查询20个
		}
		addrhash, e := fields.CheckReadableAddress(addr)
		if e != nil {
			amtstrings += "[format error],"
			continue
		}
		finditem, e1 := blcdb.Read(*addrhash)
		if e1 != nil {
			amtstrings += "[error],"
			continue
		}
		if finditem == nil {
			amtstrings += "ㄜ0:0,"
			continue
		}
		amtstrings += finditem.Amount.ToFinString() + ","
		totalamt, _ = totalamt.Add(&finditem.Amount)
	}

	// 0
	result["amounts"] = strings.TrimRight(amtstrings, ",")
	result["total"] = totalamt.ToFinString()
	return result

}
