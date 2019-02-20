package toolshell

import (
	"fmt"
	"strings"
)

var (
	mypublickey []byte

	currentInputContent string
)

/*



 */

var welcomeContent = `
Welcome to Hacash tool shell, you can:
--------
setPrivateKey("XXXX")
setPrivateKeyByPassword("XXXX")
--------
genTxSimpleTransfer("ADDRESS", "AMOUNT")
--------
quit, quit(), exit, exit()
--------
Continue to enter anything:
`

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
		params := make([]string, 0)
		switch function {
		/*****************************************************/
		case "setPrivateKey":
			setPrivateKey(params)
		case "setPrivateKeyByPassword":
			setPrivateKeyByPassword(params)
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

func setPrivateKey(params []string) {
	fmt.Println("ok, set you PrivateKey.")

}

func setPrivateKeyByPassword(params []string) {
	fmt.Println("ok, set you PrivateKey by Password.")

}
