package gentxs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"strconv"
)

// 创建一笔交易
func GenTxSimpleTransfer(ctx ctx.Context, params []string) {
	if len(params) < 4 {
		fmt.Println("params not enough")
		return
	}
	from := params[0]
	to := params[1]
	finamt := params[2]
	finfee := params[3]
	if ctx.NotLoadedYetAccountAddress(from) {
		return
	}
	toAddr := ctx.IsInvalidAccountAddress(to)
	if toAddr == nil {
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
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(fields.Address(masterAddr))
	newTrs.Timestamp = fields.VarInt5(ctx.UseTimestamp()) // 使用 hold 的时间戳
	if e5 != nil {
		fmt.Println("create transaction error, " + e5.Error())
		return
	}
	newTrs.Fee = *fee // set fee
	tranact := actions.NewAction_1_SimpleTransfer(*toAddr, *amt)
	newTrs.AppendAction(tranact)
	// sign
	e6 := newTrs.FillNeedSigns(ctx.GetAllPrivateKeyBytes(), nil)
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

	sigok, sigerr := trxnew.VerifyNeedSigns(nil)
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
	fmt.Println("hash: <" + hex.EncodeToString(newTrs.HashNoFee()) + ">, hash_with_fee: <" + hex.EncodeToString(newTrs.Hash()) + ">")
	fmt.Println("body length " + strconv.Itoa(len(bodybytes)) + " bytes, hex body is:")
	fmt.Println("-------- TRANSACTION BODY START --------")
	fmt.Println(hex.EncodeToString(bodybytes))
	//fmt.Println( hex.EncodeToString( bodybytes2 ) )
	fmt.Println("-------- TRANSACTION BODY END   --------")

	// 记录
	ctx.SetTxToRecord(newTrs.HashNoFee(), newTrs)

}
