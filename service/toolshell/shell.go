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

	acc := account.CreateAccountByPassword("HACASH+HCX+3500+161660245")

	printLoadAddress(acc)

	params := gjson.Parse("[\"11122377dCWTAyvcwMazczgPaJWsbmFomq\",\"1969418WSUCXPBSyGeLytkAqKUspDZJYWt\",\"HCX1:248\",\"HCX1:240\"]").Array()
	genTxSimpleTransfer(params)

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
