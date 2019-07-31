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
	pool2 "github.com/hacash/blockmint/miner/pool"
	"github.com/hacash/blockmint/sys/file"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// 查看所有交易
func minerPoolAllTransactions(response http.ResponseWriter, request *http.Request) {
	if len(config.Config.MiningPool.PayPassword) == 0 {
		response.Write([]byte("miner pool not open."))
		return
	}
	params := parseRequestQuery(request)
	var page_num uint64 = 0
	var transaction_id uint64 = 0
	if v, ok := params["page"]; ok {
		if num, e := strconv.ParseUint(v, 10, 0); e == nil {
			page_num = num
		}
	}
	if v, ok := params["transaction_id"]; ok {
		if num, e := strconv.ParseUint(v, 10, 0); e == nil {
			transaction_id = num
		}
	}

	// 显示数据
	htmltext := "<html><head><title>miner pool data statistics</title></head><body>"
	htmltext += `<style>#table{ border-collapse: collapse; } td{padding: 0 5px;} </style>`

	pool := miner.GetGlobalInstanceMiningPool()
	trsRec := pool.StoreDB.ReadTransferRecord(false)
	// 查看全部转账交易
	trs_end := int(trsRec.Latest - (20 * page_num))
	trs_start := trs_end - 19
	if trs_start < 1 {
		trs_start = 1
	}
	// fmt.Println("=================== page limit ", trs_start, trs_end)
	if transaction_id > 0 {
		trs_start = int(transaction_id)
		trs_end = int(transaction_id)
	} else {
		htmltext += `<h5>`
		if page_num > 0 {
			htmltext += fmt.Sprintf(`<a href="?page=%d">PrevPage</a>`, page_num-1)
		}
		if trs_start > 1 {
			htmltext += fmt.Sprintf(` · · · <a href="?page=%d">NextPage</a>`, page_num+1)
		}
		htmltext += `</h5>`
	}
	htmltext += `<table id="table" border="1">
		<tr>
			<th>#</th>
			<th>TxId</th>
			<th>Amount</th>
			<th>Address</th>
			<th>CreateTime</th>
			<th>SubmitTime</th>
			<th>Detail</th>
		</tr>
    `
	// 查询交易
	for i := trs_end; i >= trs_start; i-- {
		trs := pool.StoreDB.ReadTransfer(uint64(i))
		if trs != nil {
			time_fmt := "2006-01-02 15:04"
			detail := fmt.Sprintf(`<a target="_blank" href="?transaction_id=%d">detail</a>`, i)
			if transaction_id > 0 {
				time_fmt += ":05" // 精确到秒
				txbody := pool.StoreDB.ReadTransactionBody(trs.TxId)
				detail = `<p>Transaction Body:</p><textarea style="height:120px;width:300px;margin-bottom: 14px;">` + hex.EncodeToString(txbody) + `</textarea>`
			}
			subtime := ""
			if trs.SubmitTimestamp > 0 {
				subtime = time.Unix(int64(trs.SubmitTimestamp), 0).Format(time_fmt)
			}
			htmltext += fmt.Sprintf(`<tr>
				<td>%d</td>
				<td>%d</td>
				<td>ㄜ%d:240</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
				<td>%s</td>
			</tr>`,
				trs.Id,
				trs.TxId,
				trs.Amount,
				trs.Address.ToReadable(),
				time.Unix(int64(trs.CreateTimestamp), 0).Format(time_fmt),
				subtime,
				detail,
			)

		}
	}

	// 返回显示
	htmltext += "</table>"
	htmltext += "</body></html>"
	response.Write([]byte(htmltext))

}

// 自动打币矿池状态
func minerPoolStatisticsAutoTransfer(response http.ResponseWriter, request *http.Request) {
	if len(config.Config.MiningPool.PayPassword) == 0 {
		response.Write([]byte("miner pool not open."))
		return
	}
	// 显示数据
	htmltext := "<html><head><title>miner pool data statistics</title></head><body>"
	htmltext += `<style>#table{ border-collapse: collapse; } td{padding: 0 5px;} </style>`

	// 查看统计信息
	pool := miner.GetGlobalInstanceMiningPool()
	trsRec := pool.StoreDB.ReadTransferRecord(false)
	htmltext += fmt.Sprintf(`<div>
		<p>FeeRatio: %.2f %%</p>
		<p>Latest: %d, Submit: %d, <a href="/minerpool/transactions" target="_blank">show transactions</a></p>
		<p>TxLatestId: %d, TxConfirm: %d</p>
		<p>PrevSendHeight: %d</p>
	</div>`,
		config.Config.MiningPool.PayFeeRatio*100,
		trsRec.Latest, trsRec.Submit,
		trsRec.TxLatestId, trsRec.TxConfirm,
		trsRec.PrevSendHeight,
	)
	htmltext += `<table id="table" border="1">
		<tr>
			<th>Address</th>
			<th>RealTime-Client,Thread,Power</th>
			<th>FindBlocks,Coins</th>
			<th>CompleteRewards</th>
			<th>DeservedRewards</th>
			<th>PrevTransferBlockHeight</th>
		</tr>
    `
	// 查看当前全部在线挖矿的 矿工
	pool.StateData.AllPowWorkers.Range(func(key interface{}, val interface{}) bool {
		powwk := val.(*pool2.PowWorker)
		htmltext += fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%d , %d , %s</td>
			<td>%d , %d</td>
			<td>ㄜ%d:240</td>
			<td>ㄜ%d:240</td>
			<td>%d</td>
		</tr>`,
			powwk.RewordAddress.ToReadable(),
			powwk.ClientCount,
			powwk.RealtimeWorkSubmitCount,
			powwk.RealtimePower.String(),
			powwk.StatisticsData.FindBlocks,
			powwk.StatisticsData.FindCoins,
			powwk.StatisticsData.CompleteRewards,
			powwk.StatisticsData.DeservedRewards,
			powwk.StatisticsData.PrevTransferBlockHeight,
		)
		return true
	})

	// 返回显示
	htmltext += "</table>"
	htmltext += "</body></html>"
	response.Write([]byte(htmltext))
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

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
