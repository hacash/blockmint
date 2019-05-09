package toolshell

import (
	"fmt"
	"github.com/hacash/blockmint/service/toolshell/ctx"
	"github.com/hacash/blockmint/service/toolshell/gentxs"
)

func genTx(ctx ctx.Context, params []string) {
	if len(params) <= 1 {
		fmt.Println("params not enough")
		return
	}
	typename := params[0]
	bodys := params[1:]
	switch typename {
	case "sendcash": // 发送交易
		gentxs.GenTxSimpleTransfer(ctx, bodys)
	case "paychan": // 创建支付通道
		gentxs.GenTxCreatePaymentChannel(ctx, bodys)
	default:
		fmt.Println("Sorry, undefined gentx type: " + typename)
	}

}
