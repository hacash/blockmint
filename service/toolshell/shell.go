package toolshell

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/core/account"
	"os"
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
		"hardertodobetter4649036",   // 142159YdbweQdhqnBdXpqbyUDWEbGkhKBq
	}
	for i := 0; i < len(addrs); i++ {
		acc := account.CreateAccountByPassword(addrs[i])
		printLoadAddress(acc)
	}

	//params := gjson.Parse("[\"127717zvZWFjEghjEpyyRSnitEEbnMuuLn\",\"1AihFeuaC5xvqDP8nVcphCpvMrTR7gUdMH\",\"HCX2:244\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params)
	//params1 := gjson.Parse("[\"127717zvZWFjEghjEpyyRSnitEEbnMuuLn\",\"1969418WSUCXPBSyGeLytkAqKUspDZJYWt\",\"HCX100:244\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params1)

	//tx1 := "01005c851002000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df401030001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d540accd693d83f9dd56557a2c4db752ec3e9d0f66885da3219950faebf1e93c1b3aac4150c94bdc9610a9360b526e233aa0a76497ff47fd553e97ae87579fd2720000"
	tx2 := "01005c851012000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df401020001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d519da9900eb5fb038060e99aa1fbf9317a1e437d2c6012f002a2db44cacff01162edb31eddf0e75d78f2fc432cb9d59760b948e484752b0f445f8ef5c39b749940000"
	params3 := gjson.Parse("[\"" + tx2 + "\",\"127.0.0.1:33383\"]").Array()
	sendTxToMiner(params3)

}

func RunToolShell() {

	fmt.Println(welcomeContent)

	for true {
		fmt.Print(">>")
		inputReader := bufio.NewReader(os.Stdin)
		input, err := inputReader.ReadString('\n')
		if err != nil {
			continue
		}
		//fmt.Scanln(&currentInputContent)
		currentInputContent = strings.TrimRight(input, "\n")
		// exit
		if currentInputContent == "exit" ||
			currentInputContent == "exit()" ||
			currentInputContent == "quit" ||
			currentInputContent == "quit()" {
			fmt.Println("Bye")
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
