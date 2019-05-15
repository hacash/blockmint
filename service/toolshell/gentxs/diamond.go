package gentxs

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"github.com/hacash/x16rs"
	"strconv"
)

/*


gentx diamond $DIAMOND $NUMBER $PrevHash $Nonce $Address $feeAddress $FEE


passwd 123456
passwd 12345678
gentx diamond NHMYYM 1 000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e6 0100000001552c71 1271438866CSDpJUqrnchoJAiGGBFSQhjd 1EDUeK8NAjrgYhgDFv9NJecn8dNyJJsu3y HCX2:244

gentx diamond_transfer NHMYYM 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9 1271438866CSDpJUqrnchoJAiGGBFSQhjd HCX1:244


*/

// 创建钻石
func GenTxCreateDiamond(ctx ctx.Context, params []string) {
	if len(params) < 7 {
		fmt.Println("params not enough")
		return
	}
	diamondArgv := params[0]
	numberArgv := params[1]
	prevHashArgv := params[2]
	nonceArgv := params[3]
	addressArgv := params[4]
	feeAddressArgv := params[5]
	feeArgv := params[6]
	// 检查字段
	_, dddok := x16rs.IsDiamondHashResultString("0000000000" + diamondArgv)
	if !dddok {
		fmt.Printf("%s is not diamond value.\n", diamondArgv)
		return
	}
	number, e3 := strconv.ParseUint(numberArgv, 10, 0)
	if e3 != nil {
		fmt.Printf("number %s is error.\n", numberArgv)
		return
	}
	noncehash, e3 := hex.DecodeString(nonceArgv)
	if e3 != nil {
		fmt.Printf("nonce %s format is error.\n", nonceArgv)
		return
	}
	address := ctx.IsInvalidAccountAddress(addressArgv)
	if address == nil {
		return
	}
	feeAddress := ctx.IsInvalidAccountAddress(feeAddressArgv)
	if feeAddress == nil {
		return
	}
	feeAmount := ctx.IsInvalidAmountString(feeArgv)
	if feeAmount == nil {
		return
	}
	blkhash, e0 := hex.DecodeString(prevHashArgv)
	if e0 != nil {
		fmt.Println("block hash format error")
		return
	}
	// 创建 action
	var dimcreate actions.Action_4_DiamondCreate
	dimcreate.Number = fields.VarInt3(number)
	dimcreate.Diamond = fields.Bytes6(diamondArgv)
	dimcreate.PrevHash = blkhash
	dimcreate.Nonce = fields.Bytes8(noncehash)
	dimcreate.Address = *address
	// 创建交易
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(*feeAddress)
	newTrs.Timestamp = fields.VarInt5(ctx.UseTimestamp()) // 使用 hold 的时间戳
	if e5 != nil {
		fmt.Println("create transaction error, " + e5.Error())
		return
	}
	newTrs.Fee = *feeAmount // set fee
	// 放入action
	newTrs.AppendAction(&dimcreate)

	// 数据化
	bodybytes, e7 := newTrs.Serialize()
	if e7 != nil {
		fmt.Println("transaction serialize error, " + e7.Error())
		return
	}
	// sign
	e6 := newTrs.FillNeedSigns(ctx.GetAllPrivateKeyBytes(), nil)
	if e6 != nil {
		fmt.Println("sign transaction error, " + e6.Error())
		return
	}
	// 检查签名
	sigok, sigerr := newTrs.VerifyNeedSigns(nil)
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

/////////////////////////////////////////////////////////////////////////////////////

// 转移钻石
func GenTxDiamondTransfer(ctx ctx.Context, params []string) {
	if len(params) < 4 {
		fmt.Println("params not enough")
		return
	}
	diamondArgv := params[0]
	addressArgv := params[1]
	feeAddressArgv := params[2]
	feeArgv := params[3]
	// 检查字段
	_, dddok := x16rs.IsDiamondHashResultString("0000000000" + diamondArgv)
	if !dddok {
		fmt.Printf("%s is not diamond value.\n", diamondArgv)
		return
	}
	address := ctx.IsInvalidAccountAddress(addressArgv)
	if address == nil {
		return
	}
	feeAddress := ctx.IsInvalidAccountAddress(feeAddressArgv)
	if feeAddress == nil {
		return
	}
	feeAmount := ctx.IsInvalidAmountString(feeArgv)
	if feeAmount == nil {
		return
	}
	// 创建 action
	var dimtransfer actions.Action_5_DiamondTransfer
	dimtransfer.Diamond = fields.Bytes6(diamondArgv)
	dimtransfer.Address = *address
	// 创建交易
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(*feeAddress)
	newTrs.Timestamp = fields.VarInt5(ctx.UseTimestamp()) // 使用 hold 的时间戳
	if e5 != nil {
		fmt.Println("create transaction error, " + e5.Error())
		return
	}
	newTrs.Fee = *feeAmount // set fee
	// 放入action
	newTrs.AppendAction(&dimtransfer)

	// 数据化
	bodybytes, e7 := newTrs.Serialize()
	if e7 != nil {
		fmt.Println("transaction serialize error, " + e7.Error())
		return
	}
	// sign
	e6 := newTrs.FillNeedSigns(ctx.GetAllPrivateKeyBytes(), nil)
	if e6 != nil {
		fmt.Println("sign transaction error, " + e6.Error())
		return
	}
	// 检查签名
	sigok, sigerr := newTrs.VerifyNeedSigns(nil)
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
