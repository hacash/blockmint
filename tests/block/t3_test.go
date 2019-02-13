package block

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"testing"
)

func Test_block1(t *testing.T) {

	var testByteAry = GenTestData_trs_block()

	var block1 = new(blocks.Block_v1)

	block1.Parse(&testByteAry, 0)
	var resultByteAry, _ = block1.Serialize()

	//var bignum = block1.Amount.GetValue()

	// resultByteAry = append(resultByteAry, []byte{1}...)

	fmt.Println(testByteAry)
	fmt.Println(resultByteAry)

	fmt.Println(bytes.Equal(testByteAry, resultByteAry))
	//fmt.Print(bignum)
}
