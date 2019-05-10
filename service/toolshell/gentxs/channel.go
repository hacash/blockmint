package gentxs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"golang.org/x/crypto/sha3"
	"strconv"
)

// 创建支付通道
// gentx paychan ADDRESS1 AMOUNT1 ADDRESS2 AMOUNT2 FEE
/*

passwd 123456
passwd 12345678
gentx paychan 1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y HCX1:248 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9 HCX1:248 HCX4:244


*/
func GenTxCreatePaymentChannel(ctx ctx.Context, params []string) {
	if len(params) < 5 {
		fmt.Println("params not enough")
		return
	}
	// fmt.Println(strings.Join(params, ","))
	leftAddressArgv := params[0]
	leftAmountArgv := params[1]
	rightAddressArgv := params[2]
	rightAmountArgv := params[3]
	feeArgv := params[4]
	// 检查字段
	leftAddress := ctx.IsInvalidAccountAddress(leftAddressArgv)
	if leftAddress == nil {
		return
	}
	rightAddress := ctx.IsInvalidAccountAddress(rightAddressArgv)
	if rightAddress == nil {
		return
	}
	leftAmount := ctx.IsInvalidAmountString(leftAmountArgv)
	if leftAmount == nil {
		return
	}
	rightAmount := ctx.IsInvalidAmountString(rightAmountArgv)
	if rightAmount == nil {
		return
	}
	fee := ctx.IsInvalidAmountString(feeArgv)
	if fee == nil {
		return
	}
	// 开始拼装 action
	var paychan actions.Action_2_OpenPaymentChannel
	paychan.LeftAddress = *leftAddress
	paychan.LeftAmount = *leftAmount
	paychan.RightAddress = *rightAddress
	paychan.RightAmount = *rightAmount
	pcbts, _ := paychan.Serialize()
	bufs := bytes.NewBuffer(pcbts[16:])
	bufs.Write([]byte(strconv.FormatUint(ctx.UseTimestamp(), 10)))
	hx := sha3.Sum256(bufs.Bytes())
	paychan.ChannelId = hx[0:16]
	// 创建交易
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(*leftAddress)
	newTrs.Timestamp = fields.VarInt5(ctx.UseTimestamp()) // 使用 hold 的时间戳
	if e5 != nil {
		fmt.Println("create transaction error, " + e5.Error())
		return
	}
	newTrs.Fee = *fee // set fee
	// 放入action
	newTrs.AppendAction(&paychan)

	// 打印 hash 签名数据

	// ok
	fmt.Println("transaction create success! ")
	fmt.Println("hash: <" + hex.EncodeToString(newTrs.HashNoFee()) + ">, hash_with_fee: <" + hex.EncodeToString(newTrs.Hash()) + ">")
	fmt.Printf("( payment_channel_id = %s )\n", hex.EncodeToString(paychan.ChannelId))

	bodybytes, e7 := newTrs.Serialize()
	if e7 != nil {
		fmt.Println("transaction serialize error, " + e7.Error())
		return
	}
	fmt.Println("body length " + strconv.Itoa(len(bodybytes)) + " bytes, hex body is:")
	fmt.Println("-------- TRANSACTION BODY [NOT SIGN] START --------")
	fmt.Println(hex.EncodeToString(bodybytes))
	fmt.Println("-------- TRANSACTION BODY [NOT SIGN] END   --------")

	// 记录
	ctx.SetTxToRecord(newTrs.HashNoFee(), newTrs)

}
