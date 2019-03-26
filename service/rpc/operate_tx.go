package rpc

import (
	"encoding/hex"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/service/txpool"
	"net/http"
)

func addTxToPool(w http.ResponseWriter, value []byte) {

	defer func() {
		if err := recover(); err != nil {
			w.Write([]byte("Transaction body data error"))
		}
	}()

	var tx, _, e = blocks.ParseTransaction(value, 0)
	if e != nil {
		w.Write([]byte("Transaction format error"))
		return
	}
	//
	// 尝试加入交易池
	pool := txpool.GetGlobalInstanceMemTxPool()
	e3 := pool.AddTx(tx)
	if e3 != nil {
		w.Write([]byte("Transaction Add to MemTxPool error: \n" + e3.Error()))
		return
	}

	// ok
	hashnofee := tx.HashNoFee()
	hashnofeestr := hex.EncodeToString(hashnofee)
	w.Write([]byte("{\"success\":\"Transaction add to MemTxPool successfully !\",\"txhash\":\"" + hashnofeestr + "\"}"))

}
