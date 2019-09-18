package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	miner2 "github.com/hacash/blockmint/miner"
	"github.com/hacash/blockmint/p2p"
	txpool2 "github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/blockmint/types/block"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	dealHomePrintCacheTime  = time.Now()
	dealHomePrintCacheBytes []byte
)

func dealHome(response http.ResponseWriter, request *http.Request) {

	if len(dealHomePrintCacheBytes) > 0 && time.Now().Unix() < dealHomePrintCacheTime.Unix()+57 {
		response.Write(dealHomePrintCacheBytes)
		return
	}
	dealHomePrintCacheTime = time.Now()
	// 矿工状态
	var responseStrAry = []string{}
	var miner = miner2.GetGlobalInstanceHacashMiner()
	curheight := miner.State.CurrentHeight()
	minerblkhead := miner.State.GetBlockHead()
	prev288_90height := uint64(curheight) - (288 * 30 * 3)
	prev288_30height := uint64(curheight) - (288 * 30)
	prev288_7height := uint64(curheight) - (288 * 7)
	prev288height := uint64(curheight) / 288 * 288
	num288 := uint64(curheight) - prev288height
	if prev288_7height <= 0 {
		prev288_7height = 1
	}
	if prev288height <= 0 {
		prev288height = 1
	}
	diamondNumber, _ := miner.State.GetPrevDiamondBlockHash()
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"height: %d, tx: %d, hash: %s, difficulty: %d, create_time: %s, diamond number: %d",
		curheight,
		minerblkhead.GetTransactionCount()-1,
		hex.EncodeToString(miner.State.CurrentBlockHash()),
		minerblkhead.GetDifficulty(),
		time.Unix(int64(minerblkhead.GetTimestamp()), 0).Format("2006/01/02 15:04:05"),
		diamondNumber,
	))
	// 出块统计
	cost288_90miao := getMiao(minerblkhead, prev288_90height, 288*90)
	cost288_30miao := getMiao(minerblkhead, prev288_30height, 288*30)
	cost288_7miao := getMiao(minerblkhead, prev288_7height, 288*7)
	cost288miao := getMiao(minerblkhead, prev288height, num288)
	// fmt.Println(prev288height, num288, cost288miao)
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"block average time, last quarter: %s ( %ds/300s = %.2f), month: %s ( %ds/300s = %.4f), week: %s ( %ds/300s = %.4f), last from %d+%d: %s ( %ds/300s = %f)",
		time.Unix(int64(cost288_90miao), 0).Format("04:05"),
		cost288_90miao,
		(float32(cost288_90miao)/300),
		time.Unix(int64(cost288_30miao), 0).Format("04:05"),
		cost288_30miao,
		(float32(cost288_30miao)/300),
		time.Unix(int64(cost288_7miao), 0).Format("04:05"),
		cost288_7miao,
		(float32(cost288_7miao)/300),
		prev288height,
		num288,
		time.Unix(int64(cost288miao), 0).Format("04:05"),
		cost288miao,
		(float32(cost288miao)/300),
	))
	// 交易池信息
	txpool := txpool2.GetGlobalInstanceMemTxPool()
	if pool, ok := txpool.(*txpool2.MemTxPool); ok {
		diamonds := ""
		hd := pool.TxHead
		for i := 0; i < 200; i++ {
			if hd != nil {
				if as := hd.Tx.GetActions(); len(as) > 0 {
					if as[0].Kind() == 4 {
						if dia, ok := as[0].(*actions.Action_4_DiamondCreate); ok {
							if len(diamonds) > 0 {
								diamonds += "/" + string(dia.Diamond)
							} else {
								diamonds = string(dia.Diamond)
							}
						}
					}
				}
				hd = hd.Next
			} else {
				break
			}
		}
		responseStrAry = append(responseStrAry, fmt.Sprintf(
			"txpool length: %d, size: %fkb, diamond: %s",
			pool.Length,
			float64(pool.Size)/1024,
			diamonds,
		))
	}
	// 矿池信息
	if len(config.Config.MiningPool.StatisticsDir) > 0 {
		minerpool := miner2.GetGlobalInstanceMiningPool()
		responseStrAry = append(responseStrAry, fmt.Sprintf("miner pool connected client: %d", minerpool.StateData.ClientCount))
	}

	// 节点连接信息
	p2pserver := p2p.GetGlobalInstanceP2PServer()
	nodeinfo := p2pserver.GetServer().NodeInfo()
	p2pobj := p2p.GetGlobalInstanceProtocolManager()
	peers := p2pobj.GetPeers().PeersWithoutTx([]byte{0})
	bestpeername := ""
	for _, pr := range peers {
		bestpeername += pr.Name() + ", "
	}
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"p2p peer name: %s, enode: %s, connected: %d, connect peers: %s",
		nodeinfo.Name,
		nodeinfo.Enode,
		len(peers),
		strings.TrimRight(bestpeername, ", "),
	))

	// Write
	responseStrAry = append(responseStrAry, "")
	dealHomePrintCacheBytes = []byte("<html>" + strings.Join(responseStrAry, "\n\n<br><br> ") + "</html>")
	response.Write(dealHomePrintCacheBytes)
}

func getMiao(minerblkhead block.Block, prev288height uint64, blknum uint64) uint64 {
	storedb := store.GetGlobalInstanceBlocksDataStore()
	_, prevblockbytes, _ := storedb.GetBlockBytesByHeight(uint64(prev288height), true, false, 0)
	if len(prevblockbytes) == 0 {
		return 0
	}
	prevblock, _, _ := blocks.ParseBlockHead(prevblockbytes, 0)
	costtotalmiao := minerblkhead.GetTimestamp() - prevblock.GetTimestamp()
	if blknum == 0 {
		blknum = 1 // fix bug
	}
	costmiao := costtotalmiao / blknum
	return costmiao
}

func dealQuery(response http.ResponseWriter, request *http.Request) {
	params := parseRequestQuery(request)
	if _, ok := params["action"]; !ok {
		response.Write([]byte("must action"))
		return
	}

	// call controller
	routeQueryRequest(params["action"], params, response, request)

}

func dealOperate(response http.ResponseWriter, request *http.Request) {

	bodybytes, e1 := ioutil.ReadAll(request.Body)
	if e1 != nil {
		response.Write([]byte("body error"))
		return
	}
	if len(bodybytes) < 4 {
		response.Write([]byte("body length less than 4"))
		return
	}
	routeOperateRequest(response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func parseRequestQuery(request *http.Request) map[string]string {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
	return params
}

func RunHttpRpcService() {

	initRoutes()

	http.HandleFunc("/", dealHome)           //设置访问的路由
	http.HandleFunc("/query", dealQuery)     //设置访问的路由
	http.HandleFunc("/operate", dealOperate) //设置访问的路由

	http.HandleFunc("/minerpool", minerPoolStatisticsAutoTransfer)       //设置访问的路由
	http.HandleFunc("/minerpool/transactions", minerPoolAllTransactions) //设置访问的路由

	http.HandleFunc("/minerpool/statistics", minerPoolStatistics) //设置访问的路由

	port := config.Config.P2p.Port.Rpc

	err := http.ListenAndServe(":"+port, nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		fmt.Println("RunHttpRpcService on " + port)
	}
}
