package fields

import "bytes"

type Address string

var (
	addressMaxLen = uint32(34)
)

func (addr *Address) Serialize() ([]byte, error) {
	var str = string(*addr)
	for {
		if uint32(len(str)) < addressMaxLen {
			str += " "
		} else {
			break
		}
	}
	return []byte(str), nil
}

func (addr *Address) Parse(buf []byte, seek uint32) (uint32, error) {
	var addrbytes = buf[seek : seek+addressMaxLen]
	addrbytes = bytes.TrimRight(addrbytes, " ")
	var sd = Address(string(addrbytes))
	*addr = sd // replace
	return seek + addressMaxLen, nil
}
