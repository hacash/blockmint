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
		"hardertodobetter4649036",   // 142159YdbweQdhqnBdXpqbyUDWEbGkhKBq
	}
	for i := 0; i < len(addrs); i++ {
		acc := account.CreateAccountByPassword(addrs[i])
		printLoadAddress(acc)
	}

	//params := gjson.Parse("[\"127717zvZWFjEghjEpyyRSnitEEbnMuuLn\",\"1AihFeuaC5xvqDP8nVcphCpvMrTR7gUdMH\",\"HCX5:248\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params)
	//params1 := gjson.Parse("[\"127717zvZWFjEghjEpyyRSnitEEbnMuuLn\",\"1969418WSUCXPBSyGeLytkAqKUspDZJYWt\",\"HCX100:244\",\"HCX1:244\"]").Array()
	//genTxSimpleTransfer(params1)

	//tx1 := "01005c6fefc4000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df801010001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d5c1747e672eab9cfb909c1797bd778894c4da5b6eefff5e6d44d40a7158d7c4342fc43fbe7bb2856e2d3350e16a2d0bdb604eed36c47d9a7c7c6dfff77ba3057c0000"
	//tx1 := "01005c81edc2000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df801010001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d5065b5279a7c0160d6e95115f94a4ae02ecc5ba631226c900b205948e7eb5de50520abf982bdb12c1b2527c0e912e3d0a84e98623eb9212a7011e0330a3de49af0000"
	//tx1 := "01005c81edd5000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df801020001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d508293f31215b6f2e9aeb6ecf1613e0923ec8a593f7bd50c52ffb1e36997011022a0a26d04dd2fd69a3398d0fdac21ea1442e5bf37699c0172027f5a1cabf63470000"
	//tx1 := "01005c81ede2000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df801030001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d5cb45909f5b66656a90fad029c42e5a9d4c1e9c6f68ba5f21d9dded1973ecbcce2839fd3e17cf9d1baeef6c0cd78dfefd59cef9f3a9d9b7bbd125a11dc7033d730000"
	//tx1 := "01005c6fefc40016b3a82d6bd43c0dd145f405062a080a582ceeb3f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df901010001028879668b8d649e731bd2815142eb38b025fd6ff07544cb3380c24edc4abdca34d3a3d0dabe2a5c688232b482fe5c5d242c987b8c201c61550b991d35723fcf7b0edcb5299b2567c005402429a422050aa92f32defd64cb56a045bd32afb27be10000"
	//tx1 := "01005c81edf0000c1fa1c032d90fd7afc54deb03941e87b4c59756f4010100010001006a9bc9a70fafe1ba1b760341807af30d094bb20df801050001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d5a6eb8485d551864f18d6d27133acb971b3e961752c204ddd33d5e39f895314562a2ffa258ebc321846314e37c0db60888232aaa2fa5223a11c618a97f31cb77b0000"
	//tx2 := "01005c6fefc4000c1fa1c032d90fd7afc54deb03941e87b4c59756f40101000100010058b9ceaea0e0bd4cfcca96ef2cec052234a5e6d3f601010001039ffe91e6b39c21c32d282272ce83ce852d021cd34044181e94dad754b0f7c7d569346d85c3e589c05fd8c18249b811bed76096e81ebcf13e43c67960d1922a0422fe26f886794c1933b6ff3648ac6bb034a047842de7d6a8fa33ad942783a4ff0000"
	//params3 := gjson.Parse("[\"" + tx1 + "\",\"49.51.34.40:3338\"]").Array()
	//sendTxToMiner(params3)

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
