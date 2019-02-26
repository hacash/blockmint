package fields

import (
	//"fmt"
	//"unsafe"
	"bytes"
	"github.com/hacash/blockmint/sys/err"
)

var EmptyZeroBytes32 = bytes.Repeat([]byte{0}, 32)
var EmptyZeroBytes512 = bytes.Repeat([]byte{0}, 512)

type Bytes21 []byte
type Bytes32 []byte
type Bytes33 []byte
type Bytes64 []byte

////////////////////////////////////////////////////////

func (elm *Bytes21) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 21) }
func (elm *Bytes32) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 32) }
func (elm *Bytes33) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 33) }
func (elm *Bytes64) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 64) }

func (elm *Bytes21) Parse(buf []byte, seek uint32) (uint32, error) {
	return bytesParse(elm, buf, seek, 21)
}
func (elm *Bytes32) Parse(buf []byte, seek uint32) (uint32, error) {
	return bytesParse(elm, buf, seek, 32)
}
func (elm *Bytes33) Parse(buf []byte, seek uint32) (uint32, error) {
	return bytesParse(elm, buf, seek, 33)
}
func (elm *Bytes64) Parse(buf []byte, seek uint32) (uint32, error) {
	return bytesParse(elm, buf, seek, 64)
}

func (elm *Bytes21) Size() uint32 { return 21 }
func (elm *Bytes32) Size() uint32 { return 32 }
func (elm *Bytes33) Size() uint32 { return 33 }
func (elm *Bytes64) Size() uint32 { return 64 }

////////////////////////////////////////////////////////

func bytesParse(elm interface{}, buf []byte, seek uint32, maxlen uint32) (uint32, error) {
	var addrbytes = buf[seek : seek+maxlen]
	//var sd = string(addrbytes)
	switch a := elm.(type) {
	case *Bytes21:
		*a = (Bytes21(addrbytes))
	case *Bytes32:
		*a = (Bytes32(addrbytes))
	case *Bytes33:
		*a = (Bytes33(addrbytes))
	case *Bytes64:
		*a = (Bytes64(addrbytes))
	default:
		return 0, err.New("not find type")
	}
	//elm = sd // replace
	return seek + maxlen, nil
}

func bytesSerialize(str string, maxlen uint32) ([]byte, error) {
	//var str = string(*elm)
	for {
		if uint32(len(str)) < maxlen {
			str += " "
		} else {
			break
		}
	}
	return []byte(str), nil
}
