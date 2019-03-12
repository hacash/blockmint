package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	miner2 "github.com/hacash/blockmint/miner"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func dealHome(response http.ResponseWriter, request *http.Request) {
	// 矿工状态
	var responseStrAry = []string{}
	var miner = miner2.GetGlobalInstanceHacashMiner()
	curheight := miner.State.CurrentHeight()
	minerblkhead := miner.State.GetBlockHead()
	prevheight := int64(curheight) - (288 * 7)
	if prevheight <= 0 {
		prevheight = 1
	}
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"height: %d, tx: %d, hash: %s, difficulty: %d, create_time: %s",
		curheight,
		minerblkhead.GetTransactionCount()-1,
		hex.EncodeToString(miner.State.CurrentBlockHash()),
		minerblkhead.GetDifficulty(),
		time.Unix(int64(minerblkhead.GetTimestamp()), 0).Format("2006/01/02 15:04:05"),
	))
	// 出块统计
	prevblockbytes, _ := store.GetGlobalInstanceBlocksDataStore().GetBlockBytesByHeight(uint64(prevheight), true, false)
	prevblock, _, _ := blocks.ParseBlockHead(prevblockbytes, 0)
	costtotalmiao := minerblkhead.GetTimestamp() - prevblock.GetTimestamp()
	costmiao := costtotalmiao / (288 * 7)
	prevblock.GetTimestamp()
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"last week block average time: %s ( %d / 300 = %f S)",
		time.Unix(int64(costmiao), 0).Format("04:05"),
		costmiao,
		(float32(costmiao)/300),
	))

	// Write
	responseStrAry = append(responseStrAry, "")
	response.Write([]byte("<html>" + strings.Join(responseStrAry, "\n\n<br><br> ") + "</html>"))
}

func dealQuery(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
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
	}
	routeOperateRequest(response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func RunHttpRpcService() {

	initRoutes()

	http.HandleFunc("/", dealHome)           //设置访问的路由
	http.HandleFunc("/query", dealQuery)     //设置访问的路由
	http.HandleFunc("/operate", dealOperate) //设置访问的路由

	port := config.Config.P2p.Port.Rpc

	err := http.ListenAndServe(":"+port, nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		fmt.Println("RunHttpRpcService on " + port)
	}
}
