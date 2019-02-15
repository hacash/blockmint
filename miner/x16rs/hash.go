package x16rs

import "github.com/hacash/x16rs"

func HashX16RS(stuff *[]byte) []byte {
	return x16r.HashX16RS(*stuff)
}
