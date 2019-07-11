package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/account"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/sys/file"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

func minerPoolStatistics(response http.ResponseWriter, request *http.Request) {
	params := parseRequestQuery(request)
	if _, ok := params["dataid"]; !ok {
		response.Write([]byte("dataid is must"))
		return
	}
	var payratio float64 = 0.0 // 矿池支付比例
	if payratio_param, ok := params["payratio"]; ok {
		//fmt.Println(payratio_param)
		if num, err := strconv.ParseFloat(payratio_param, 0); err == nil {
			//fmt.Println(payratio)
			payratio = num
		}
	}
	// 付款之账号
	var payAccount *account.Account = nil
	if password, ok := params["paypassword"]; ok {
		//fmt.Println(password)
		payAccount = account.CreateAccountByPassword(password)
	}
	// 手续费
	fee, _ := fields.NewAmountFromFinString("HCX1:244")
	if feestrcover, ok := params["payfee"]; ok {
		f, e := fields.NewAmountFromFinString(feestrcover)
		if e == nil {
			fee = f
		}
	}
	// 时间
	txusetime := time.Now()
	if _, ok := params["paytime"]; ok {

	}

	// 读取文件
	filename := config.Config.MiningPool.StatisticsDir + "/miningpooldata.dat." + params["dataid"]
	if !file.IsExist(filename) {
		response.Write([]byte("data file is not exist"))
		return
	}
	// 解析数据
	dataobj := miner.ReadStoreDataByFileName(filename)

	// 显示数据
	htmltext := "<html><head><title> pool data statistics </title></head><body>"

	htmltext += fmt.Sprintf("<h5>Rewards: ㄜ%d:248, Address: %d, TotalPower: %s",
		dataobj.SuccessRewards, dataobj.NodeNumber, dataobj.TotalMiningPower.String())
	if payAccount != nil {
		htmltext += `, PayAddress: ` + string(payAccount.AddressReadable)
	}
	htmltext += `</h5>`
	htmltext += `<style>#table{ border-collapse: collapse; } td{padding: 0 5px;} </style>`
	htmltext += `<table id="table" border="1">
		<tr>
			<th>#</th>
			<th>Address</th>
			<th>Power</th>
			<th>Ratio</th>
			<th>Rewards</th>
    `
	if payratio > 0 {
		htmltext += `<th>Payout</th>` // 向矿工支付
		if payAccount != nil {
			htmltext += `<th>TxHash</th><th>TransferBody</th>` // 支付交易
		}
	}
	htmltext += `
		</tr>
	`
	index := 0
	dataobj.MiningPowerStatistics.Range(func(key interface{}, val interface{}) bool {
		index++
		addrstr := key.(string)
		value := val.(*big.Int)
		//client.Close()
		_, err := fields.CheckReadableAddress(addrstr)
		if err == nil {
			rationum := new(big.Float).Quo(new(big.Float).SetInt(dataobj.TotalMiningPower), new(big.Float).SetInt(value))
			basemu, _ := rationum.Float64()
			rwdz := 1.0 / basemu * float64(dataobj.SuccessRewards)
			htmltext += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>ㄜ%.8f:248</td>",
				index,
				addrstr,
				value.String(),
				"1/"+rationum.String(),
				rwdz,
			)
			if payratio > 0 {
				realpay := rwdz * payratio
				realpayint := int64(realpay * 100000000)
				htmltext += fmt.Sprintf(`<td>ㄜ%.8f:248</td>`, realpay) // 向矿工支付
				realpayintstr := fmt.Sprintf(`ㄜ%d:240`, realpayint)
				if payAccount != nil { // 创建转账交易并签名
					tx, txbody := createPoolWorkerPayTx(payAccount, addrstr, realpayintstr, fee, txusetime.Unix())
					htmltext += `<td><input value="` + tx + `"/></td><td><textarea>` + txbody + `</textarea></td>`
				}
			}
			htmltext += `</tr>`
		}
		return true
	})

	htmltext += "</table>"
	htmltext += "</body></html>"

	response.Write([]byte(htmltext))

}

// 创建转账交易
func createPoolWorkerPayTx(payacc *account.Account, toaddr string, amount_str string, fee *fields.Amount, unix int64) (string, string) {
	//fmt.Println(toaddr, amount_str)
	newTrs, e1 := transactions.NewEmptyTransaction_2_Simple(payacc.Address)
	toaddress, e2 := fields.CheckReadableAddress(toaddr)
	amount, e3 := fields.NewAmountFromFinString(amount_str)
	//fmt.Println(e1, e2, e3)
	if e1 != nil || e2 != nil || e3 != nil {
		return "error", "create error"
	}
	newTrs.Timestamp = fields.VarInt5(unix) // 使用时间戳
	newTrs.Fee = *fee
	tranact := actions.NewAction_1_SimpleTransfer(*toaddress, *amount)
	newTrs.AppendAction(tranact)
	// 私钥
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(payacc.Address)] = payacc.PrivateKey
	// 签名
	e9 := newTrs.FillNeedSigns(allPrivateKeyBytes, nil)
	if e9 != nil {
		return "error", "sign error"
	}
	txbody, e10 := newTrs.Serialize()
	if e10 != nil {
		return "error", "serialize body error"
	}
	// 返回创建的交易字符串
	return hex.EncodeToString(newTrs.HashNoFeeFresh()), hex.EncodeToString(txbody)
}
