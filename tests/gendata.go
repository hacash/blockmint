package tests

import "bytes"

func GenTestData_block() []byte {

	var testbuffer bytes.Buffer

	// action 1  head
	testbuffer.Write([]byte{1})                                  // version
	testbuffer.Write([]byte{0, 0, 0, 0, 8})                      // height
	testbuffer.Write([]byte{0, 0, 0, 0, 9})                      // timestamp
	testbuffer.Write([]byte("oooooooooooooooooooooooooooooooo")) // prevMark 32
	testbuffer.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")) // mrklRoot 32
	testbuffer.Write([]byte{0, 0, 0, 2})                         // transactionCount
	// head end

	testbuffer.Write([]byte{1, 2, 3, 4}) // miner nonce

	testbuffer.Write(GenTestData_trs_coinbase())
	testbuffer.Write(GenTestData_trs_simple())

	var testByteAry = testbuffer.Bytes()

	return testByteAry
}

func GenTestData_trs_coinbase() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0})                       // kind
	testbuffer.Write([]byte("addrass0000000addrass")) // address
	testbuffer.Write([]byte{248, 1, 1})               // reward
	testbuffer.Write([]byte("########        "))      // message hardertodobetter

	return testbuffer.Bytes()
}

func GenTestData_trs_simple() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{1})                       // kind
	testbuffer.Write([]byte{99, 99, 99, 99, 99})      // Timestamp
	testbuffer.Write([]byte("addrass1111111addrass")) // address
	testbuffer.Write([]byte{248, 1, 77})              // fee
	testbuffer.Write([]byte{1})                       // actionCount

	testbuffer.Write(GenTestData_action_transfer())

	return testbuffer.Bytes()
}

func GenTestData_action_transfer() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0, 1})                    // kind
	testbuffer.Write([]byte("addrass2222222addrass")) // addrass
	testbuffer.Write([]byte{248, 1, 88})              // amount

	return testbuffer.Bytes()
}
