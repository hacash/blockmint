package rpc

import (
	"encoding/json"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/store"
	"net/http"
)

var (
	queryRoutes = make(map[string]func(map[string]string) map[string]string)
)

func initRoutes() {
	queryRoutes["balance"] = getBalance // 查询余额

}

func routeQueryRequest(action string, params map[string]string, w http.ResponseWriter, r *http.Request) {
	if ctrl, ok := queryRoutes[action]; ok {
		resobj := ctrl(params)
		restxt, e1 := json.Marshal(resobj)
		if e1 != nil {
			w.Write([]byte("data not json"))
		} else {
			w.Write(restxt)
		}
	} else {
		w.Write([]byte("not find action"))
	}
}

//////////////////////////////////////////////////////////////

func getBalance(params map[string]string) map[string]string {
	result := make(map[string]string)
	addr, ok1 := params["address"]
	if !ok1 {
		result["err"] = "address must"
		return result
	}
	addrhash, e := fields.CheckReadableAddress(addr)
	if e != nil {
		result["err"] = e.Error()
		return result
	}
	blcdb := store.GetGlobalInstanceChainStateBalanceDB()
	finditem, e1 := blcdb.Read(*addrhash)
	if e1 != nil {
		result["err"] = "find error"
		return result
	}
	if finditem != nil {
		result["amount"] = finditem.Amount.ToAccountingString()
		return result
	}

	// 0
	result["amount"] = fields.AmountToZeroAccountingString()
	return result

}

//////////////////////////////////////////////////////////////
