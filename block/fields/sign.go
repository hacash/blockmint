package fields

import (
	"bytes"
)

//////////////////////////////////////////////////////////////////

type Sign struct {
	PublicKey Bytes33
	Signature Bytes64
}

func (this *Sign) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(this.PublicKey)
	buffer.Write(this.Signature)
	return buffer.Bytes(), nil
}

func (this *Sign) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = this.PublicKey.Parse(buf, seek)
	seek, _ = this.Signature.Parse(buf, seek)
	return seek, nil
}

func (this *Sign) Size() uint32 {
	return this.PublicKey.Size() + this.Signature.Size()
}

//////////////////////////////////////////////////////////////////

type Multisign struct {
	CondElem      uint8 // 分子
	CondBase      uint8 // 分母
	PublicKeyList []Bytes33
	SignatureInds []uint8
	SignatureList []Bytes64
}

func (this *Multisign) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{this.CondElem, this.CondElem})
	length1 := int(this.CondElem)
	length2 := int(this.CondBase)
	for i := 0; i < length2; i++ {
		buffer.Write(this.PublicKeyList[i])
	}
	for j := 0; j < length1; j++ {
		buffer.Write([]byte{this.SignatureInds[j]})
	}
	for k := 0; k < length1; k++ {
		buffer.Write(this.PublicKeyList[k])
	}
	return buffer.Bytes(), nil
}

func (this *Multisign) Parse(buf []byte, seek uint32) (uint32, error) {
	this.CondElem = buf[seek]
	this.CondBase = buf[seek+1]
	seek = seek + 2
	length1 := int(this.CondElem)
	length2 := int(this.CondBase)
	this.PublicKeyList = make([]Bytes33, length2)
	this.SignatureInds = make([]uint8, length1)
	this.SignatureList = make([]Bytes64, length1)
	var e error
	for i := 0; i < length2; i++ {
		var b Bytes33
		seek, e = b.Parse(buf, seek)
		if e != nil {
			return 0, e
		}
		this.PublicKeyList[i] = b
		seek += b.Size()
	}
	for i := 0; i < length1; i++ {
		this.SignatureInds[i] = buf[seek]
		seek += 1
	}
	for i := 0; i < length1; i++ {
		var b Bytes64
		seek, e = b.Parse(buf, seek)
		if e != nil {
			return 0, e
		}
		this.SignatureList[i] = b
		seek += b.Size()
	}
	return seek, nil
}

func (this *Multisign) Size() uint32 {
	length := uint32(this.CondBase)
	return 1 + 1 + length*33 + length*1 + length*64
}
