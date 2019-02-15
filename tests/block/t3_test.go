package block

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/tests"
	"testing"
)

func Test_block1(t *testing.T) {

	var testByteAry = tests.GenTestData_block()

	var block1, _, _ = blocks.ParseBlock(testByteAry, 0)
	var resultByteAry, _ = block1.Serialize()

	//var bignum = block1.Amount.GetValue()

	// resultByteAry = append(resultByteAry, []byte{1}...)

	fmt.Println(testByteAry)
	fmt.Println(resultByteAry)

	fmt.Println(bytes.Equal(testByteAry, resultByteAry))
	//fmt.Print(bignum)
}
