package toolshell

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/core/account"
	"github.com/hacash/blockmint/types/block"
	"os"
	"strings"
	"time"
)

var (
	MyAccounts          = make(map[string]account.Account, 0)
	AllPrivateKeyBytes  = make(map[string][]byte, 0)
	Transactions        = make(map[string]block.Transaction, 0)
	TargetTime          time.Time // 使用的时间
	currentInputContent string
)

/*

 */

type ctxToolShell struct{}

func (ctxToolShell) NotLoadedYetAccountAddress(addr string) bool {
	if _, ok := MyAccounts[addr]; !ok {
		fmt.Println("Account " + addr + " need to be loaded")
		return true
	}
	return false
}

func (ctxToolShell) IsInvalidAccountAddress(addr string) bool {
	if _, err := base58check.Decode(addr); err != nil {
		return true
	}
	return false
}

func (ctxToolShell) GetAllPrivateKeyBytes() map[string][]byte {
	return AllPrivateKeyBytes
}

func (ctxToolShell) SetTxToRecord(hash_no_fee []byte, tx block.Transaction) { // 记录交易
	Transactions[string(hash_no_fee)] = tx
}

func (ctxToolShell) GetTxFromRecord(hash_no_fee []byte) block.Transaction { // 获取交易
	if tx, ok := Transactions[string(hash_no_fee)]; ok {
		return tx
	} else {
		fmt.Println("Not find tx " + hex.EncodeToString(hash_no_fee))
		return nil
	}
}

func (ctxToolShell) UseTimestamp() uint64 { // 当前使用的时间戳
	return uint64(TargetTime.Unix())
}

////////////////////////////////////

var welcomeContent = `
Welcome to Hacash tool shell, you can:
--------
passwd $XXX $XXX  |  prikey $0xAB123D...  |  newkey  |  accounts  |  update
--------
gentx sendcash $FROM_ADDRESS $TO_ADDRESS $AMOUNT $FEE  |  loadtx $0xTXBODYBYTES  |  txs
--------
sendtx $TXHASH $IP:PORT
--------
exit, quit
--------`

func RunToolShell() {

	fmt.Println(welcomeContent)
	TargetTime = time.Now()
	fmt.Println("The use time hold on " + TargetTime.Format("2006/01/02 15:04:05") + ", enter 'update' change to now")
	fmt.Println("Continue to enter anything:")

	inputReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">>")
		input, err := inputReader.ReadString('\n')
		if err != nil {
			continue
		}
		//fmt.Scanln(&currentInputContent)
		currentInputContent = strings.TrimSpace(input)
		// empty
		if currentInputContent == "" {
			continue
		}
		// exit
		if currentInputContent == "exit" ||
			currentInputContent == "quit" {
			fmt.Println("Bye")
			break
		}
		if currentInputContent == "update" {
			TargetTime = time.Now()
			fmt.Println("Hold time change to " + TargetTime.Format("2006/01/02 15:04:05"))
			continue
		}
		if currentInputContent == "accounts" {
			showAccounts()
			continue
		}
		if currentInputContent == "txs" {
			showTxs()
			continue
		}
		// some opration
		params := strings.Fields(currentInputContent)
		funcname := params[0]
		parabody := params[1:]
		switch params[0] {
		case "passwd":
			setPrivateKeyByPassword(parabody)
		case "prikey":
			setPrivateKey(parabody)
		case "newkey":
			createNewPrivateKey(parabody)
		case "gentx":
			genTx(ctxToolShell{}, parabody)
		case "sendtx":
			sendTxToMiner(ctxToolShell{}, parabody)
		default:
			fmt.Println("Sorry, undefined instructions: " + funcname)
		}
		// clear
		currentInputContent = ""
	}
}

/////////////////////////////////////////////////

/////////////////////////////////////////////////////////

func showAccounts() {
	if len(MyAccounts) == 0 {
		fmt.Println("none")
		return
	}
	for k, _ := range MyAccounts {
		fmt.Print(k + " ")
	}
	fmt.Print("\n")
}

func showTxs() {
	if len(Transactions) == 0 {
		fmt.Println("none")
		return
	}
	for k, _ := range Transactions {
		fmt.Println(hex.EncodeToString([]byte(k)))
	}
}

func setPrivateKey(params []string) {
	for _, hexstr := range params {
		if strings.HasPrefix(hexstr, "0x") {
			hexstr = string([]byte(hexstr)[2:]) // drop 0x
		}
		_, e0 := hex.DecodeString(hexstr)
		if e0 != nil {
			fmt.Println("Private Key '" + hexstr + "' is error")
			return
		}
		acc, e1 := account.GetAccountByPriviteKeyString(hexstr)
		if e1 != nil {
			fmt.Println("Private Key '" + hexstr + "' is error")
			return
		}
		printLoadAddress(acc)
	}
}

func setPrivateKeyByPassword(params []string) {
	for _, passwd := range params {
		//fmt.Println(passwd)
		acc := account.CreateAccountByPassword(passwd)
		printLoadAddress(acc)
	}
}

// 随机创建私钥
func createNewPrivateKey(params []string) {
	acc := account.CreateNewAccount()
	printLoadAddress(acc)
}

func printLoadAddress(acc *account.Account) {
	MyAccounts[string(acc.AddressReadable)] = *acc // append
	AllPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
	fmt.Println("Ok, has loaded your account private key: 0x" + hex.EncodeToString(acc.PrivateKey) + " address: " + string(acc.AddressReadable))
}
