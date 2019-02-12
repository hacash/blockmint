package fields

import (
	"bytes"
	//"encoding/binary"
	"math/big"
)

type Amount struct {
	Dist    int8
	Numeral []byte
	Unit    uint8
}

func (bill *Amount) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{byte(bill.Dist)})
	buffer.Write(bill.Numeral)
	buffer.Write([]byte{bill.Unit})
	return buffer.Bytes(), nil
}

func (bill *Amount) Parse(buf *[]byte, seek int) (int, error) {
	bill.Dist = int8((*buf)[seek])
	var numCount = int(bill.Dist)
	if bill.Dist < 0 {
		numCount *= -1
	}
	var tailstart = seek + 1 + numCount
	bill.Numeral = (*buf)[seek+1 : tailstart]
	bill.Unit = uint8((*buf)[tailstart])

	return tailstart + 1, nil
}

func (bill *Amount) GetValue() *big.Int {
	var bignum = new(big.Int)
	bignum.SetBytes(bill.Numeral)
	var sign = big.NewInt(int64(big.NewInt(int64(bill.Dist)).Sign()))
	var unit = new(big.Int)
	unit.Exp(big.NewInt(int64(10)), big.NewInt(int64(bill.Unit)), big.NewInt(int64(0)))
	bignum.Mul(bignum, unit)
	bignum.Mul(bignum, sign) // do sign
	return bignum
}
