package fields

import (
	"bytes"
	"github.com/hacash/bitcoin/address/base58check"
	"github.com/hacash/blockmint/sys/err"
)

type AddressReadable = TrimString34
type Address = Bytes21

// 检查地址合法性
func CheckReadableAddress(readable string) (*Address, error) {
	if len(readable) > 34 {
		return nil, err.New("Address format error")
	}
	hashhex, e1 := base58check.Decode(readable)
	if e1 != nil {
		return nil, err.New("Address format error")
	}
	version := uint8(hashhex[0])
	if version > 2 {
		return nil, err.New("Address version error")
	}
	addr := Address(hashhex)
	return &addr, nil
}

// 有效的地址
func (this *Address) IsValid() bool {
	if this == nil {
		return false
	}
	if len(*this) != 21 {
		return false
	}
	if 0 == bytes.Compare(*this, bytes.Repeat([]byte{0}, 21)) {
		return false
	}
	// ok
	return true
}
