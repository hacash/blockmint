package tests

import (
	"bytes"
)

func GenTestData_block_set_height(height uint32) []byte {

	var testbuffer bytes.Buffer

	// action 1  head
	testbuffer.Write([]byte{1})                                                                                                           // version
	testbuffer.Write([]byte{0, uint8(height / (256 * 256 * 256)), uint8(height / (256 * 256)), uint8(height / 256), uint8(height % 256)}) // height
	testbuffer.Write([]byte{0, 0, 0, 0, 9})                                                                                               // timestamp
	testbuffer.Write([]byte("oooooooooooooooooooooooooooooooo"))                                                                          // prevMark 32
	testbuffer.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))                                                                          // mrklRoot 32
	testbuffer.Write([]byte{0, 0, 0, 2})                                                                                                  // transactionCount
	// head end

	testbuffer.Write([]byte{1, 2, 3, 4}) // miner nonce
	testbuffer.Write([]byte{6, 7, 8, 9}) // Difficulty
	testbuffer.Write([]byte{0, 0})       // WitnessStage

	// transactions
	testbuffer.Write(GenTestData_trs_coinbase())
	testbuffer.Write(GenTestData_trs_simple())

	var testByteAry = testbuffer.Bytes()

	return testByteAry

}

func GenTestData_block() []byte {
	return GenTestData_block_set_height(8)
}

func GenTestData_trs_coinbase() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0})                       // kind
	testbuffer.Write([]byte("addrass0000000addrass")) // address
	testbuffer.Write([]byte{248, 1, 1})               // reward
	testbuffer.Write([]byte("########        "))      // message hardertodobetter
	testbuffer.Write([]byte{0})                       // witness count

	return testbuffer.Bytes()
}

func GenTestData_trs_simple() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{1})                       // kind
	testbuffer.Write([]byte{99, 99, 99, 99, 99})      // CreateTimestamp
	testbuffer.Write([]byte("addrass1111111addrass")) // address
	testbuffer.Write([]byte{248, 1, 77})              // fee
	testbuffer.Write([]byte{0, 1})                    // actionCount

	testbuffer.Write(GenTestData_action_transfer())

	testbuffer.Write([]byte{0, 0}) // SignCount
	testbuffer.Write([]byte{0, 0}) // MultiSignCount

	return testbuffer.Bytes()
}

func GenTestData_action_transfer() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0, 1})                    // kind
	testbuffer.Write([]byte("addrass2222222addrass")) // addrass
	testbuffer.Write([]byte{248, 1, 88})              // amount

	return testbuffer.Bytes()
}
