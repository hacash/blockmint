package x16rs

import (
	"github.com/hacash/x16rs"
)

func HashX16RS(loopnum int, stuff []byte) []byte {
	//fmt.Println( "=> " + hex.EncodeToString(stuff) )
	res := x16rs.HashX16RS(loopnum, stuff)
	//fmt.Println( "<= " + hex.EncodeToString(res) + " =>" )
	return res
}

func HashX16RS_Optimize(loopnum int, stuff []byte) []byte {
	//fmt.Println( "=> " + hex.EncodeToString(stuff) )
	res := x16rs.HashX16RS_Optimize(loopnum, stuff)
	//fmt.Println( "<= " + hex.EncodeToString(res) + " =>" )
	return res
}
