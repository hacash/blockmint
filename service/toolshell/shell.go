package toolshell

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/core/account"
	"strings"

	"github.com/tidwall/gjson"
)

var (
	MyAccounts         = make(map[string]account.Account, 0)
	AllPrivateKeyBytes = make(map[string][]byte, 0)

	currentInputContent string
)

/*



 */

var welcomeContent = `
Welcome to Hacash tool shell, you can:
--------
setPrivateKey("XXXX") | setPrivateKeyByPassword("XXXX") | showAccounts()
--------
genTxSimpleTransfer("FROM ADDRESS", "TO ADDRESS", "AMOUNT", "FEE")
--------
sendTxToMiner("TXBODY", "MINER_ADDR")
--------
quit, quit(), exit, exit()
--------
Continue to enter anything:
`

func RunTest() {

	addrs := []string{
		"hardertodobetter1116350",   // 127717zvZWFjEghjEpyyRSnitEEbnMuuLn
		"hardertodobetter156439106", // 1969418WSUCXPBSyGeLytkAqKUspDZJYWt
		"hardertodobetter2363390",   // 135361CpCMxbfLEEPdVrmudJKQnQpPKZJv
	}
	for i := 0; i < len(addrs); i++ {
		acc := account.CreateAccountByPassword(addrs[i])
		printLoadAddress(acc)
	}

	//params := gjson.Parse("[\"127717zvZWFjEghjEpyyRSnitEEbnMuuLn\",\"1969418WSUCXPBSyGeLytkAqKUspDZJYWt\",\"HCX1:248\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params)
	//params := gjson.Parse("[\"1969418WSUCXPBSyGeLytkAqKUspDZJYWt\",\"135361CpCMxbfLEEPdVrmudJKQnQpPKZJv\",\"HCX1:244\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params)

	//tx1 := "01005c6fefc4000c1fa1c032d90fd7afc54deb03941e87b4c59756f40101000100010058b9ceaea0e0bd4cfcca96ef2cec052234a5e6d3f801010001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d5a887a3a45bf5652259b6538ef546adef58a101ec25a3cd4e3729383e1430307f79ab68398e97b82296f78e36e54f66b009fb04e9f0774fe1d21ba8f6438ae7890000"
	tx2 := "01005c6fefc40058b9ceaea0e0bd4cfcca96ef2cec052234a5e6d3f40101000100010016b3a82d6bd43c0dd145f405062a080a582ceeb3f40101000102bc40e9ca301b81e4aa914465098ba43b065d3d4e88d809c8a6fdf81294117a4bb30988e477a37219898e00c93d902d0d83097fe80cb102ccb779969bf44f937114d9fc6130647ceb900b77775caeedecb19ee3741e2df5e7f75aa8d0656aa1ef0000"
	params2 := gjson.Parse("[\"" + tx2 + "\",\"127.0.0.1:3338\"]").Array()
	sendTxToMiner(params2)

}

func RunToolShell() {

	fmt.Println(welcomeContent)

	for true {
		fmt.Print(">>")
		fmt.Scanln(&currentInputContent)
		// exit
		if currentInputContent == "exit" ||
			currentInputContent == "exit()" ||
			currentInputContent == "quit" ||
			currentInputContent == "quit()" {
			fmt.Println("Bye.")
			break
		}
		// empty
		if currentInputContent == "" {
			continue
		}
		// route call
		fcnx := strings.IndexByte(currentInputContent, byte('('))
		if fcnx == -1 {
			errorFormat()
			continue
		}
		function := currentInputContent[0:fcnx]
		paramsptr := parseParams(currentInputContent[fcnx:])
		if paramsptr == nil {
			errorFormat()
			continue
		}
		params := *paramsptr
		switch function {
		/*****************************************************/
		case "showAccounts":
			showAccounts(params)
		case "setPrivateKey":
			setPrivateKey(params)
		case "setPrivateKeyByPassword":
			setPrivateKeyByPassword(params)
		case "genTxSimpleTransfer":
			genTxSimpleTransfer(params)
		case "sendTxToMiner":
			sendTxToMiner(params)
		/*****************************************************/
		default:
			fmt.Println("Sorry, undefined call: " + function + "(...)")
		}

		currentInputContent = "" // clear
	}
}

/////////////////////////////////////////////////

func errorFormat() {
	fmt.Println("Ops! call format error.")
	currentInputContent = "" // clear
}

func parseParams(stuff string) *[]gjson.Result {
	if strings.HasPrefix(stuff, "(") && strings.HasSuffix(stuff, ")") {
		stuffbytes := []byte(stuff)
		sflen := len(stuffbytes)
		if sflen == 2 {
			res := make([]gjson.Result, 0)
			return &res // empty param
		}
		stuffbytes = stuffbytes[1 : sflen-1]
		//fmt.Println(string(stuffbytes))
		params := gjson.Parse("[" + string(stuffbytes) + "]").Array()
		//fmt.Println(params[0].String())
		return &params
	}
	return nil
}

/////////////////////////////////////////////////////////

func showAccounts(params []gjson.Result) {
	if len(MyAccounts) == 0 {
		fmt.Println("none")
		return
	}
	for k, _ := range MyAccounts {
		fmt.Println(k)
	}
}

func setPrivateKey(params []gjson.Result) {
	hexstr := params[0].String()
	_, e0 := hex.DecodeString(hexstr)
	if e0 != nil {
		fmt.Println("Private Key error")
		return
	}
	acc, e1 := account.GetAccountByPriviteKeyString(hexstr)
	if e1 != nil {
		fmt.Println("Private Key error")
		return
	}
	printLoadAddress(acc)
}

func setPrivateKeyByPassword(params []gjson.Result) {
	passwd := params[0].String()
	//fmt.Println(passwd)
	acc := account.CreateAccountByPassword(passwd)
	printLoadAddress(acc)
}

func printLoadAddress(acc *account.Account) {
	MyAccounts[string(acc.AddressReadable)] = *acc // append
	AllPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
	fmt.Println("Ok, has loaded your account, address is:\n" + acc.AddressReadable)
}

func notLoadedYetAccountAddress(addr string) bool {
	if _, ok := MyAccounts[addr]; !ok {
		fmt.Println("Account " + addr + " need to be loaded")
		return true
	}
	return false
}

func isInvalidAccountAddress(addr string) bool {
	if _, err := base58check.Decode(addr); err != nil {
		return true
	}
	return false
}
