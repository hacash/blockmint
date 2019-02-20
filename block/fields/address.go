package fields

import (
	"encoding/hex"
	"github.com/anaskhan96/base58check"
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
	addrhash, _ := hex.DecodeString(hashhex)
	version := uint8(addrhash[0])
	if version > 2 {
		return nil, err.New("Address version error")
	}
	addr := Address(addrhash)
	return &addr, nil
}
