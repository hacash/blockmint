package toolshell

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/tidwall/gjson"
)

// 创建一笔交易
func genTxSimpleTransfer(params []gjson.Result) {
	if len(params) < 3 {
		fmt.Println("params not enough")
		return
	}
	from := params[0].String()
	to := params[1].String()
	finamt := params[2].String()
	finfee := params[3].String()
	if notLoadedYetAccountAddress(from) {
		return
	}
	if isInvalidAccountAddress(to) {
		return
	}
	amt, e1 := fields.NewAmountFromFinString(finamt)
	if e1 != nil {
		fmt.Println("amount format error or over range, the right example is 'HCX1:248' for one coin")
		return
	}
	fee, e2 := fields.NewAmountFromFinString(finfee)
	if e2 != nil {
		fmt.Println("fee format error or over range")
		return
	}
	masterAddr, e3 := base58check.Decode(from)
	if e3 != nil {
		fmt.Println("from address format error")
		return
	}
	toAddr, e4 := base58check.Decode(to)
	if e4 != nil {
		fmt.Println("from address format error")
		return
	}
	newTrs, e5 := transactions.NewEmptyTransaction_1_Simple(fields.Address(masterAddr))
	if e5 != nil {
		fmt.Println("create transaction error, " + e5.Error())
		return
	}
	newTrs.Fee = *fee // set fee
	tranact := actions.NewAction_1_SimpleTransfer(toAddr, *amt)
	newTrs.AppendAction(tranact)
	// sign
	e6 := newTrs.FillNeedSigns(AllPrivateKeyBytes)
	if e6 != nil {
		fmt.Println("sign transaction error, " + e6.Error())
		return
	}
	bodybytes, e7 := newTrs.Serialize()
	if e7 != nil {
		fmt.Println("transaction serialize error, " + e7.Error())
		return
	}

	var trxnew, _, _ = blocks.ParseTransaction(bodybytes, 0)
	bodybytes2, _ := trxnew.Serialize()
	if 0 != bytes.Compare(bodybytes, bodybytes2) {
		fmt.Println("transaction serialize error")
		return
	}

	sigok, sigerr := trxnew.VerifyNeedSigns()
	if sigerr != nil {
		fmt.Println("transaction VerifyNeedSigns error")
		return
	}
	if !sigok {
		fmt.Println("transaction VerifyNeedSigns fail")
		return
	}

	// ok
	fmt.Println("transaction create success! ")
	fmt.Println("hash: <" + hex.EncodeToString(newTrs.Hash()) + ">")
	fmt.Println("hash_no_fee: <" + hex.EncodeToString(newTrs.HashNoFee()) + ">")
	fmt.Println("the transaction hex body is:")
	fmt.Println("-------- TRANSACTION BODY START --------")
	fmt.Println(hex.EncodeToString(bodybytes))
	//fmt.Println( hex.EncodeToString( bodybytes2 ) )
	fmt.Println("-------- TRANSACTION BODY END   --------")

}
