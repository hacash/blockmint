package toolshell

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"strconv"
)

func getTx(ctx ctx.Context, params []string) {
	if len(params) < 1 {
		fmt.Println("params not enough")
		return
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

	// 判断是否完成签名
	sigok, sigerr := newTrs.VerifyNeedSigns()
	nosigntip := ""
	if !sigok || sigerr != nil {
		nosigntip = " [NOT SIGN]"
		fmt.Println("Attention: transaction verify need signs fail!")
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
