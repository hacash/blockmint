package rpc

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
)

//////////////////////////////////////////////////////////////

func getDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	dmstr, ok1 := params["name"]
	if !ok1 {
		result["err"] = "params name must."
		return result
	}
	if len(dmstr) != 6 {
		result["fail"] = "name format error."
		return result
	}

	dmdb := db.GetGlobalInstanceDiamondDB()
	addr, e1 := dmdb.Read(fields.Bytes6(dmstr))
	if e1 != nil || addr == nil {
		result["fail"] = "not find."
		return result
	}
	// 0
	result["address"] = addr.ToReadable()
	return result
}
