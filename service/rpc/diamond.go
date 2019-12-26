package rpc

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"strconv"
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
	store, e1 := dmdb.Read(fields.Bytes6(dmstr))
	if e1 != nil || store == nil {
		result["fail"] = "not find."
		return result
	}
	// 0
	result["block_height"] = strconv.FormatUint(uint64(store.BlockHeight), 10)
	result["number"] = strconv.Itoa(int(store.Number))
	result["address"] = store.Address.ToReadable()
	return result
}
