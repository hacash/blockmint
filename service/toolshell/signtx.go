package toolshell

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"strconv"
)

func signTx(ctx ctx.Context, params []string) {
	if len(params) < 1 {
		fmt.Println("params not enough")
		return
	}
	var adln = len(params) - 1
	var addresslist = make([][]byte, 0, adln)
	for i := 1; i < len(params); i++ {
		address := ctx.IsInvalidAccountAddress(params[i])
		if address == nil {
			return
		}
		addresslist = append(addresslist, *address)
	}
	txhashnofee, err := hex.DecodeString(params[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	newTrs := ctx.GetTxFromRecord(txhashnofee)
	if newTrs == nil {
		fmt.Printf(" tx <%s> not find!", params[0])
		return
	}
	// ok
	fmt.Println("hash: <" + hex.EncodeToString(newTrs.HashNoFee()) + ">, hash_with_fee: <" + hex.EncodeToString(newTrs.Hash()) + ">")

	// 执行签名
	// sign  // 并且加入新增的需要签名的数据
	e6 := newTrs.FillNeedSigns(ctx.GetAllPrivateKeyBytes(), addresslist)
	if e6 != nil {
		fmt.Println("sign transaction error: " + e6.Error())
		return
	}

	// 判断是否完成签名
	sigok, sigerr := newTrs.VerifyNeedSigns(nil)
	nosigntip := ""
	if !sigok || sigerr != nil {
		nosigntip = " [NOT SIGN COMPLETELY]"
		fmt.Println("Attention: transaction verify need signs fail!", sigerr)
		return
	}

	bodybytes, e7 := newTrs.Serialize()
	if e7 != nil {
		fmt.Println("transaction serialize error, " + e7.Error())
		return
	}
	fmt.Println("body length " + strconv.Itoa(len(bodybytes)) + " bytes, hex body is:")
	fmt.Println("-------- TRANSACTION BODY" + nosigntip + " START --------")
	fmt.Println(hex.EncodeToString(bodybytes))
	fmt.Println("-------- TRANSACTION BODY" + nosigntip + " END   --------")

	// 记录
	ctx.SetTxToRecord(newTrs.HashNoFee(), newTrs)

}
