package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	queryRoutes = make(map[string]func(map[string]string) map[string]string)
)

func initRoutes() {
	queryRoutes["balance"] = getBalance          // 查询余额
	queryRoutes["passwd"] = newAccountByPassword // 通过密码创建账户
	queryRoutes["createtx"] = transferSimple     // 创建普通转账交易
	queryRoutes["txconfirm"] = txStatus          // 查询交易确认状态
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
