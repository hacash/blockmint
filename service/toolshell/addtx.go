package toolshell

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

// 发送一笔交易给矿工
func sendTxToMiner(params []gjson.Result) {
	if len(params) < 2 {
		fmt.Println("params not enough")
		return
	}
	txbody := params[0].String()
	minerAddress := params[1].String()

	txbytes, e0 := hex.DecodeString(txbody)
	if e0 != nil {
		fmt.Println("Tx body hex error")
		return
	}
	tx, _, e1 := blocks.ParseTransaction(txbytes, 0)
	if e1 != nil {
		fmt.Println("Tx body error")
		return
	}
	sigok, e2 := tx.VerifyNeedSigns()
	if e2 != nil || !sigok {
		fmt.Println("Tx sign error")
		fmt.Println(e2)
		return
	}
	// post 发送
	body := new(bytes.Buffer)
	body.Write([]byte{0, 0, 0, 1}) // opcode
	body.Write(txbytes)
	req, e3 := http.NewRequest("POST", "http://"+minerAddress+"/operate", body)
	if e3 != nil {
		fmt.Println("POST NewRequest error:")
		fmt.Println(e3)
		return
	}
	client := &http.Client{}
	resp, e4 := client.Do(req)
	if e4 != nil {
		fmt.Println("POST client.Do(req) error:")
		fmt.Println(e4)
		return
	}
	defer resp.Body.Close()

	// ok
	fmt.Println("add tx to " + minerAddress + ", the response is:")
	resbody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(resbody))
}
