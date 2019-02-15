package fields

import (
	"bytes"
	"math/big"
)

type Amount struct {
	Unit    uint8
	Dist    int8
	Numeral []byte
}

func (bill *Amount) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{bill.Unit})
	buffer.Write([]byte{byte(bill.Dist)})
	buffer.Write(bill.Numeral)
	return buffer.Bytes(), nil
}

func (bill *Amount) Parse(buf []byte, seek uint32) (uint32, error) {
	bill.Unit = uint8(buf[seek])
	bill.Dist = int8(buf[seek+1])
	var numCount = int(bill.Dist)
	if bill.Dist < 0 {
		numCount *= -1
	}
	var tail = seek + 2 + uint32(numCount)
	bill.Numeral = buf[seek+2 : tail]
	return tail, nil
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
