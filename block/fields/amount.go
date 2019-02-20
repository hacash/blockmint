package fields

import (
	"bytes"
	"math/big"
	"strconv"
)

type Amount struct {
	Unit    uint8
	Dist    int8
	Numeral []byte
}

func NewAmountByBigInt(bignum *big.Int) *Amount {
	longnumstrary := []byte(bignum.String())
	strlen := len(longnumstrary)
	unit := 0
	for i := strlen - 1; i >= 0; i-- {
		if string(longnumstrary[i]) == "0" {
			unit++
			if unit == 255 {
				break
			}
		} else {
			break
		}
	}
	numeralstr := string(longnumstrary[0 : strlen-unit])
	numeralbigint, ok1 := new(big.Int).SetString(numeralstr, 10)
	if !ok1 {
		panic("Amount too big")
	}
	numeralbytes := numeralbigint.Bytes()
	dist := len(numeralbytes)
	if dist > 127 {
		panic("Amount too big")
	}
	return &Amount{
		Unit:    uint8(unit),
		Dist:    int8(dist),
		Numeral: numeralbytes,
	}
	return nil
}

func NewAmount(unit uint8, num []byte) *Amount {
	dist := len(num)
	if dist > 127 {
		panic("Amount Numeral too long !")
	}
	return &Amount{
		Unit:    unit,
		Dist:    int8(dist),
		Numeral: num,
	}
}

func NewAmountSmall(num uint8, unit uint8) *Amount {
	return &Amount{
		Unit:    unit,
		Dist:    1,
		Numeral: []byte{num},
	}
}

func (bill *Amount) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
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

func (bill *Amount) Size() uint32 {
	return 1 + 1 + uint32(len(bill.Numeral))
}

//////////////////////////////////////////////////////////

func (bill *Amount) Copy() *Amount {
	num := make([]byte, len(bill.Numeral))
	copy(num, bill.Numeral)
	return &Amount{
		Unit:    bill.Unit,
		Dist:    bill.Dist,
		Numeral: num,
	}
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

func AmountToZeroAccountingString() string {
	return "ㄜ0:0"
}

func (bill *Amount) ToAccountingString() string {
	unitStr := strconv.Itoa(int(bill.Unit)) // string(bytes.Repeat([]byte{48}, int(bill.Unit)))
	numStr := new(big.Int).SetBytes(bill.Numeral).String()
	sig := ""
	if bill.Dist < 0 {
		sig = "-"
	}
	return "ㄜ" + sig + numStr + ":" + unitStr
}

// 省略小数部分 为了存进 23 位空间里面
func (bill *Amount) EllipsisDecimalFor23SizeStore() *Amount {
	maxnumlen := 23 - 1 - 1
	if len(bill.Numeral) <= maxnumlen {
		return bill
	}
	// 省略小数部分
	longnumstr := new(big.Int).SetBytes(bill.Numeral).String()
	baselen := 0
	mvseek := len(longnumstr) / 2
	for true {
		longnumcut := string([]byte(longnumstr)[0 : baselen+mvseek])
		cutnum, ok := new(big.Int).SetString(longnumcut, 10)
		if !ok {
			panic("Amount to big !!!")
		}
		mvseek = mvseek / 2
		if mvseek == 0 {
			mvseek = 1 // 最小移动
		}
		cutbytes := cutnum.Bytes()
		if len(cutbytes) == maxnumlen {
			sig := int8(1)
			if bill.Dist < 0 {
				sig = -1
			}
			unit := int(bill.Unit) + (len(longnumstr) - len(longnumcut))
			if unit > 255 {
				panic("Amount to big !!!")
			}
			return &Amount{
				uint8(unit),
				19 * sig,
				cutbytes,
			}
		} else if len(cutbytes) > maxnumlen {
			baselen -= mvseek

		} else if len(cutbytes) < maxnumlen {
			baselen += mvseek
		}
	}
	panic("Amount Ellipsis Decimal Error")
	return nil
}

// 加法
func (bill *Amount) Add(amt *Amount) *Amount {
	add1 := bill.GetValue()
	add2 := amt.GetValue()
	add1 = add1.Add(add1, add2)
	return NewAmountByBigInt(add1)
}

// 减法
func (bill *Amount) Sub(amt *Amount) *Amount {
	add1 := bill.GetValue()
	add2 := amt.GetValue()
	add1 = add1.Sub(add1, add2)
	return NewAmountByBigInt(add1)
}

// 比较
func (bill *Amount) LessThan(amt *Amount) bool {
	add1 := bill.GetValue()
	add2 := amt.GetValue()
	res := add1.Cmp(add2)
	if res == -1 {
		return true
	} else {
		return false
	}
}
