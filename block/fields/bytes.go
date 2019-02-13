package fields

import (
	//"fmt"
	//"unsafe"
	"github.com/hacash/blockmint/sys/err"
)

type Bytes32 []byte
type Bytes64 []byte

////////////////////////////////////////////////////////

func (elm *Bytes32) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 32) }
func (elm *Bytes64) Serialize() ([]byte, error) { return bytesSerialize(string(*elm), 64) }

func (elm *Bytes32) Parse(buf *[]byte, seek int) (int, error) { return bytesParse(elm, buf, seek, 32) }
func (elm *Bytes64) Parse(buf *[]byte, seek int) (int, error) { return bytesParse(elm, buf, seek, 64) }

////////////////////////////////////////////////////////

func bytesSerialize(str string, maxlen int) ([]byte, error) {
	//var str = string(*elm)
	for {
		if len(str) < maxlen {
			str += " "
		} else {
			break
		}
	}
	return []byte(str), nil
}

func bytesParse(elm interface{}, buf *[]byte, seek int, maxlen int) (int, error) {
	var addrbytes = (*buf)[seek : seek+maxlen]
	//var sd = string(addrbytes)
	switch a := elm.(type) {
	case *Bytes32:
		*a = (Bytes32(addrbytes))
	case *Bytes64:
		*a = (Bytes64(addrbytes))
	default:
		//fmt.Println("")
		return 0, err.New("not find type")
	}
	//elm = sd // replace
	return seek + maxlen, nil
}
