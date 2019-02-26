package rpc

import (
	"encoding/hex"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/service/txpool"
	"net/http"
)

func addTxToPool(w http.ResponseWriter, value []byte) {

	var tx, _, e = blocks.ParseTransaction(value, 0)
	if e != nil {
		w.Write([]byte("Transaction format error"))
		return
	}
	ok, e1 := tx.VerifyNeedSigns()
	if !ok || e1 != nil {
		w.Write([]byte("Transaction Verify Signature error"))
		return
	}
	hashnofee := tx.HashNoFee()
	hashnofeestr := hex.EncodeToString(hashnofee)
	//
	stoblk := store.GetGlobalInstanceBlocksDataStore()
	//fmt.Println("CheckTransactionExist: " + hashnofeestr)
	ext, e2 := stoblk.CheckTransactionExist(hashnofee)
	if e2 != nil {
		w.Write([]byte("Transaction CheckTransactionExist error"))
		return
	}
	if ext {
		w.Write([]byte("Transaction " + hashnofeestr + " already exist"))
		return
	}
	// 最低手续费
	if tx.FeePurity() < uint64(config.MinimalFeePurity) {
		w.Write([]byte("The handling fee is too low for miners to accept."))
		return
	}
	// 加入交易池
	pool := txpool.GetGlobalInstanceMemTxPool()
	e3 := pool.AddTx(tx)
	if e3 != nil {
		w.Write([]byte("Transaction Add to MemTxPool error: " + e3.Error()))
		return
	}

	// ok
	w.Write([]byte("Transaction " + hashnofeestr + " Add to MemTxPool success !"))

}
