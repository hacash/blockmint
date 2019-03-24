package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"net/http"
	"strings"
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

func routeOperateRequest(w http.ResponseWriter, opcode uint32, value []byte) {
	switch opcode {
	/////////////////////////////
	case 1:
		addTxToPool(w, value)
	/////////////////////////////
	default:
		w.Write([]byte(fmt.Sprint("not find opcode %d", opcode)))
	}
}

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
	for _, addr := range addrs {
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

//////////////////////////////////////////////////////////////
