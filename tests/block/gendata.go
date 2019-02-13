package block

import "bytes"

func GenTestData_trs_block() []byte {

	var testbuffer bytes.Buffer

	// action 1
	testbuffer.Write([]byte{1})             // version
	testbuffer.Write([]byte{0, 2, 3, 4, 5}) // height
	testbuffer.Write([]byte{0, 0, 0, 0, 9}) // timestamp

	testbuffer.Write([]byte("oooooooooooooooooooooooooooooo00")) // prevMark 32
	testbuffer.Write([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx00")) // mrklRoot 32

	testbuffer.Write([]byte{0, 0, 0, 2}) // transactionCount

	testbuffer.Write(GenTestData_trs_coinbase())
	testbuffer.Write(GenTestData_trs_simple())

	var testByteAry = testbuffer.Bytes()

	return testByteAry
}

func GenTestData_trs_coinbase() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0})                                    // kind
	testbuffer.Write([]byte("29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM")) // address
	testbuffer.Write([]byte{1, 1, 248})                            // reward
	testbuffer.Write([]byte("########        "))                   // message hardertodobetter

	return testbuffer.Bytes()
}

func GenTestData_trs_simple() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{1})                                    // kind
	testbuffer.Write([]byte{99, 99, 99, 99, 99})                   // Timestamp
	testbuffer.Write([]byte("29aqbMhiK6F2s53gNp2ghoT4EezFFPpXuM")) // address
	testbuffer.Write([]byte{1, 77, 248})                           // fee
	testbuffer.Write([]byte{1})                                    // actionCount

	testbuffer.Write(GenTestData_action_transfer())

	return testbuffer.Bytes()
}

func GenTestData_action_transfer() []byte {
	var testbuffer bytes.Buffer

	testbuffer.Write([]byte{0, 1})                                 // kind
	testbuffer.Write([]byte("1313Rta8Ce99H7N5iKbGq7xp13BbAdQHmD")) // addrass
	testbuffer.Write([]byte{1, 88, 248})                           // amount

	return testbuffer.Bytes()
}
