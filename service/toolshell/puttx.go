package toolshell

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/service/toolshell/ctx"
)

func putTx(ctx ctx.Context, params []string) {
	if len(params) < 1 {
		fmt.Println("params not enough")
		return
	}
	txbody, err := hex.DecodeString(params[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	// 解析交易
	newTrs, _, err2 := blocks.ParseTransaction(txbody, 0)
	if err2 != nil {
		fmt.Println(err2)
		return
	}

	// 交易加入

	// ok
	fmt.Println("transaction append success! ")
	fmt.Println("hash: <" + hex.EncodeToString(newTrs.HashNoFee()) + ">, hash_with_fee: <" + hex.EncodeToString(newTrs.Hash()) + ">")

	// 记录
	ctx.SetTxToRecord(newTrs.HashNoFee(), newTrs)

}
