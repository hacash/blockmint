package fields

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/sys/err"
	"unsafe"
	//"unsafe"
)

type VarInt1 uint8
type VarInt2 uint16
type VarInt3 uint32
type VarInt4 uint32
type VarInt5 uint64
type VarInt6 uint64
type VarInt7 uint64
type VarInt8 uint64

////////////////////////////////////////////////////////

func (elm *VarInt1) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 1) }
func (elm *VarInt2) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 2) }
func (elm *VarInt3) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 3) }
func (elm *VarInt4) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 4) }
func (elm *VarInt5) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 5) }
func (elm *VarInt6) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 6) }
func (elm *VarInt7) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 7) }
func (elm *VarInt8) Serialize() ([]byte, error) { return varIntSerialize(uint64(*elm), 8) }

func (elm *VarInt1) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 1)
}
func (elm *VarInt2) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 2)
}
func (elm *VarInt3) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 3)
}
func (elm *VarInt4) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 4)
}
func (elm *VarInt5) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 5)
}
func (elm *VarInt6) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 6)
}
func (elm *VarInt7) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 7)
}
func (elm *VarInt8) Parse(buf []byte, seek uint32) (uint32, error) {
	return varIntParse(elm, buf, seek, 8)
}

func (elm *VarInt1) Size() uint32 { return 1 }
func (elm *VarInt2) Size() uint32 { return 2 }
func (elm *VarInt3) Size() uint32 { return 3 }
func (elm *VarInt4) Size() uint32 { return 4 }
func (elm *VarInt5) Size() uint32 { return 5 }
func (elm *VarInt6) Size() uint32 { return 6 }
func (elm *VarInt7) Size() uint32 { return 7 }
func (elm *VarInt8) Size() uint32 { return 8 }

////////////////////////////////////////////////////////

func varIntSerialize(val uint64, maxlen uint32) ([]byte, error) {
	var intbytes = bytes.Repeat([]byte{0}, 8)
	binary.BigEndian.PutUint64(intbytes, val)
	byyy := intbytes[8-maxlen : 8]
	//fmt.Println(intbytes)
	//fmt.Println("---- %d", maxlen)
	//fmt.Println(byyy)
	return byyy, nil
}

func varIntParse(elm interface{}, buf []byte, seek uint32, maxlen uint32) (uint32, error) {
	// fmt.Println("xxx",*buf)
	nnnold := buf[seek : seek+maxlen]
	var intbytes = make([]byte, len(nnnold))
	copy(intbytes, nnnold)
	// fmt.Println(intbytes)
	padbytes := bytes.Repeat([]byte{0}, int(8-maxlen))
	intbytes = append(padbytes, intbytes...)
	//addrbytes = bytes.TrimRight(addrbytes, " ")
	val := binary.BigEndian.Uint64(intbytes)
	// fmt.Println(intbytes)
	// fmt.Println("====== %d", val)
	switch a := elm.(type) {
	case *VarInt1:
		// v:= (val)>>56
		// fmt.Println("**** %d", v)
		*a = *(*VarInt1)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt2:
		// v:= val>>48
		// fmt.Println("**** %d", v)
		*a = *(*VarInt2)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt3:
		// v:= val>>48
		// fmt.Println("**** %d", v)
		*a = *(*VarInt3)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt4:
		// v:= val>>32
		// fmt.Println("**** %d", v)
		*a = *(*VarInt4)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt5:
		*a = *(*VarInt5)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt6:
		*a = *(*VarInt6)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt7:
		*a = *(*VarInt7)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	case *VarInt8:
		*a = *(*VarInt8)(unsafe.Pointer(&val))
		// fmt.Println("------- %d", *a)
	default:
		//fmt.Println("")
		return 0, err.New("not find type")
	}

	return seek + maxlen, nil
}
