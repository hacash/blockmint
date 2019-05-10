package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/service/txpool"
	"net/http"
)

func addTxToPool(w http.ResponseWriter, value []byte) {

	fmt.Println(hex.EncodeToString(value))

	defer func() {
		if err := recover(); err != nil {
			w.Write([]byte("Transaction body data error "))
		}
	}()

	fmt.Println("---- 1 ----")

	var tx, _, e = blocks.ParseTransaction(value, 0)
	if e != nil {
		w.Write([]byte("Transaction format error"))
		return
	}
	fmt.Println("---- 2 ----")
	txbts, _ := tx.Serialize()
	fmt.Println(hex.EncodeToString(txbts))
	//
	// 尝试加入交易池
	pool := txpool.GetGlobalInstanceMemTxPool()
	fmt.Println("---- 3 ----")
	e3 := pool.AddTx(tx)
	fmt.Println("---- 4 ----")
	if e3 != nil {
		w.Write([]byte("Transaction Add to MemTxPool error: \n" + e3.Error()))
		return
	}

	fmt.Println("---- 5 ----")
	// ok
	hashnofee := tx.HashNoFee()
	hashnofeestr := hex.EncodeToString(hashnofee)
	w.Write([]byte("{\"success\":\"Transaction add to MemTxPool successfully !\",\"txhash\":\"" + hashnofeestr + "\"}"))

}
