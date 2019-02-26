package x16rs

import (
	"github.com/hacash/x16rs"
)

func HashX16RS(stuff []byte) []byte {
	//fmt.Println( "=> " + hex.EncodeToString(stuff) )
	res := x16r.HashX16RS(stuff)
	//fmt.Println( "<= " + hex.EncodeToString(res) + " =>" )
	return res
}
