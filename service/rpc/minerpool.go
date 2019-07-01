package rpc

import (
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/sys/file"
	"math/big"
	"net/http"
)

func minerPoolStatistics(response http.ResponseWriter, request *http.Request) {
	params := parseRequestQuery(request)
	if _, ok := params["dataid"]; !ok {
		response.Write([]byte("dataid is must"))
		return
	}
	filename := config.Config.MiningPool.StatisticsDir + "/miningpooldata.dat." + params["dataid"]
	if !file.IsExist(filename) {
		response.Write([]byte("data file is not exist"))
		return
	}
	// 解析数据
	dataobj := miner.ReadStoreDataByFileName(filename)

	// 显示数据
	htmltext := "<html><head><title> pool data statistics </title></head><body>"

	htmltext += fmt.Sprintf("<h5>Rewards: ㄜ%d:248, Address: %d, TotalPower: %s</h5>",
		dataobj.SuccessRewards, dataobj.NodeNumber, dataobj.TotalMiningPower.String())

	htmltext += `<style>#table{ border-collapse: collapse; } td{padding: 0 5px;} </style>`
	htmltext += `<table id="table" border="1">
		<tr>
			<th>#</th>
			<th>Address</th>
			<th>Power</th>
			<th>Ratio</th>
			<th>Rewards</th>
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
			htmltext += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>ㄜ%.8f:248</td></tr>",
				index,
				addrstr,
				value.String(),
				"1/"+rationum.String(),
				1.0/basemu*float64(dataobj.SuccessRewards),
			)
		}
		return true
	})

	htmltext += "</table>"
	htmltext += "</body></html>"

	response.Write([]byte(htmltext))

}
