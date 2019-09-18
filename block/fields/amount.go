package fields

import (
	"bytes"
	"github.com/hacash/blockmint/sys/err"
	"math/big"
	"strconv"
	"strings"
)

type Amount struct {
	Unit    uint8
	Dist    int8
	Numeral []byte
}

func ParseAmount(buf []byte, seek uint32) *Amount {
	empty := NewEmptyAmount()
	empty.Parse(buf, seek)
	return empty
}

func NewEmptyAmount() *Amount {
	return &Amount{
		Unit:    0,
		Dist:    0,
		Numeral: []byte{},
	}
}

func NewAmountNumSmallCoin(num uint8) *Amount {
	return &Amount{
		Unit:    248,
		Dist:    1,
		Numeral: []byte{num},
	}
}

func NewAmountNumOneByUnit(unit uint8) *Amount {
	return &Amount{
		Unit:    unit,
		Dist:    1,
		Numeral: []byte{1},
	}
}

func NewAmountByBigIntWithUnit(bignum *big.Int, unit int) (*Amount, error) {
	var unitint = new(big.Int).Exp(big.NewInt(int64(10)), big.NewInt(int64(unit)), big.NewInt(int64(0)))
	//fmt.Println(bignum.String())
	//fmt.Println(unitint.String())
	return NewAmountByBigInt(bignum.Mul(bignum, unitint))
}

func NewAmountByBigInt(bignum *big.Int) (*Amount, error) {
	longnumstr := bignum.String()
	if longnumstr == "0" {
		return NewEmptyAmount(), nil
	}
	longnumstrary := []byte(longnumstr)
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
	//fmt.Println("longnumstrary:", bignum.String())
	//fmt.Println("numeralstr:", numeralstr)
	numeralbigint, ok1 := new(big.Int).SetString(numeralstr, 10)
	if !ok1 {
		return nil, err.New("Amount too big")
	}
	numeralbytes := numeralbigint.Bytes()
	dist := len(numeralbytes)
	if dist > 127 {
		return nil, err.New("Amount too big")
	}
	return &Amount{
		Unit:    uint8(unit),
		Dist:    int8(dist),
		Numeral: numeralbytes,
	}, nil
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
	var nnnold = buf[seek+2 : tail]
	bill.Numeral = make([]byte, len(nnnold))
	copy(bill.Numeral, nnnold)
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
	var unit = new(big.Int).Exp(big.NewInt(int64(10)), big.NewInt(int64(bill.Unit)), big.NewInt(int64(0)))
	bignum.Mul(bignum, unit)
	bignum.Mul(bignum, sign) // do sign
	return bignum
}

func (bill *Amount) IsEmpty() bool {
	return bill.Dist == int8(0) || len(bill.Numeral) == 0
}

// 判断必须为正数，且不能为零
func (bill *Amount) IsPositive() bool {
	if bill.Unit == 0 {
		return false
	}
	if bill.Dist <= 0 {
		return false
	}
	// 满足要求
	return true
}

func AmountToZeroFinString() string {
	return "ㄜ0:0"
}

// 从记账单位创建
func NewAmountFromFinString(finstr string) (*Amount, error) {
	finstr = strings.ToUpper(finstr)
	finstr = strings.Replace(finstr, " ", "", -1)
	var sig = 1
	if strings.HasPrefix(finstr, "HCX") {
		finstr = string([]byte(finstr)[3:])
	}
	if strings.HasPrefix(finstr, "ㄜ") {
		finstr = string([]byte(finstr)[3:])
	}
	if strings.HasPrefix(finstr, "-") {
		finstr = string([]byte(finstr)[1:])
		sig = -1 // 负数
	}
	var main, dum, unit string
	var main_num, dum_num *big.Int
	var unit_num int
	var e error
	var ok bool
	part := strings.Split(finstr, ":")
	if len(part) != 2 {
		return nil, err.New("format error")
	}
	unit = part[1]
	unit_num, e = strconv.Atoi(unit)
	if e != nil {
		return nil, err.New("format error")
	}
	if unit_num < 0 || unit_num > 255 {
		return nil, err.New("format error")
	}
	part2 := strings.Split(part[0], ":")
	if len(part2) < 1 || len(part2) > 2 {
		return nil, err.New("format error")
	}

	main = part2[0]
	main_num, ok = new(big.Int).SetString(main, 10)
	if !ok {
		return nil, err.New("format error")
	}
	if len(part2) == 2 {
		dum = part2[1]
		dum_num, ok = new(big.Int).SetString(dum, 10)
		if !ok {
			return nil, err.New("format error")
		}
	}
	// 处理小数部分
	bigint0, _ := new(big.Int).SetString("0", 10)
	bigint1, _ := new(big.Int).SetString("1", 10)
	bigint10, _ := new(big.Int).SetString("10", 10)
	dum_wide10 := 0
	if dum_num != nil && dum_num.Cmp(bigint0) == 1 {
		mover := dum_num.Div(dum_num, bigint10).Add(dum_num, bigint1)
		dum_wide10 = int(mover.Int64())
		if unit_num-dum_wide10 < 0 {
			return nil, err.New("format error")
		}
		main_num = main_num.Sub(main_num, mover)
		unit_num = unit_num - int(dum_wide10)
	}
	// 负数
	if sig == -1 {
		main_num = main_num.Neg(main_num)
	}
	// 转换
	return NewAmountByBigIntWithUnit(main_num, unit_num)
}

func (bill *Amount) ToFinString() string {
	unitStr := strconv.Itoa(int(bill.Unit)) // string(bytes.Repeat([]byte{48}, int(bill.Unit)))
	numStr := new(big.Int).SetBytes(bill.Numeral).String()
	sig := ""
	if bill.Dist < 0 {
		sig = "-"
	}
	return "ㄜ" + sig + numStr + ":" + unitStr
}

// 省略小数部分 为了存进 23 位空间里面
func (bill *Amount) EllipsisDecimalFor23SizeStore() (*Amount, bool) {
	maxnumlen := 23 - 1 - 1
	if len(bill.Numeral) <= maxnumlen {
		return bill, false
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
			}, true // 改变了
		} else if len(cutbytes) > maxnumlen {
			baselen -= mvseek

		} else if len(cutbytes) < maxnumlen {
			baselen += mvseek
		}
	}
	panic("Amount Ellipsis Decimal Error")
	return nil, false
}

// 加法
func (bill *Amount) Add(amt *Amount) (*Amount, error) {
	num1 := bill.GetValue()
	num2 := amt.GetValue()
	num1 = num1.Add(num1, num2)
	return NewAmountByBigInt(num1)
}

// 减法
func (bill *Amount) Sub(amt *Amount) (*Amount, error) {
	num1 := bill.GetValue()
	num2 := amt.GetValue()
	num1 = num1.Sub(num1, num2)
	return NewAmountByBigInt(num1)
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

// 相等
func (bill *Amount) Equal(amt *Amount) bool {
	if bill.GetValue().Cmp(amt.GetValue()) == 0 {
		return true
	}
	return false
}
