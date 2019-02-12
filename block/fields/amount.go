package fields

type Amount struct {
	Dist    int8
	Numeral []byte
	Unit    uint8
}

func (bill *Amount) Serialize() ([]byte, error) {
	return bill.Numeral, nil
}

func (bill *Amount) Parse(buf *[]byte, seek int) int {
	bill.Dist = int8((*buf)[seek])
	var numCount = int(bill.Dist)
	if bill.Dist < 0 {
		numCount *= -1
	}
	var tailstart = seek + 1 + numCount
	bill.Numeral = (*buf)[seek+1 : tailstart]
	bill.Unit = uint8((*buf)[tailstart])

	return tailstart + 1
}
