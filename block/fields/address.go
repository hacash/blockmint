package fields

import "bytes"

type Address string

var (
	addressMaxLen = 34
)

func (addr *Address) Serialize() ([]byte, error) {
	var str = string(*addr)
	for {
		if len(str) < addressMaxLen {
			str += " "
		} else {
			break
		}
	}
	return []byte(str), nil
}

func (addr *Address) Parse(buf *[]byte, seek int) (int, error) {
	var addrbytes = (*buf)[seek : seek+addressMaxLen]
	addrbytes = bytes.TrimRight(addrbytes, " ")
	var sd = Address(string(addrbytes))
	*addr = sd // replace
	return seek + addressMaxLen, nil
}
