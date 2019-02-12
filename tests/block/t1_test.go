package block

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/actions"
	"testing"
)

func Test_action1(t *testing.T) {
	var testbuffer bytes.Buffer

	// action 1
	//testbuffer.Write( []byte{ 0, 1 } )
	testbuffer.Write([]byte("1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD"))
	testbuffer.Write([]byte{1, 123, 248})

	var testByteAry = testbuffer.Bytes()
	var act1 = new(actions.Action_1_SimpleTransfer)

	act1.Parse(&testByteAry, 0)

	fmt.Print(testByteAry)

	var resultByteAry, _ = act1.Serialize()

	var bignum = act1.Amount.GetValue()

	fmt.Print(resultByteAry)
	fmt.Print(bignum)
}

/**
123
00000000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000
00000000000000000000000000000000000000000000000000
000000000000000000000000000000000000000000000000
*/
