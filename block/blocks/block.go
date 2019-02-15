package blocks

import (
	"fmt"
	"github.com/hacash/blockmint/sys/err"
	typesblock "github.com/hacash/blockmint/types/block"
)

func NewBlockByVersion(ty uint8) (typesblock.Block, error) {
	switch ty {
	////////////////////  TRANSATION  ////////////////////
	case 1:
		return new(Block_v1), nil
		////////////////////     END      ////////////////////
	}
	return nil, err.New("Cannot find Transaction type of " + string(ty))
}

func ParseBlock(buf []byte, seek uint32) (typesblock.Block, uint32, error) {
	version := uint8(buf[seek])
	var blk, _ = NewBlockByVersion(version)
	var mv, err = blk.Parse(buf, seek+1)
	return blk, mv, err
}

func ParseBlockHead(buf []byte, seek uint32) (typesblock.Block, uint32, error) {
	version := uint8(buf[seek])
	var blk, ee = NewBlockByVersion(version)
	if ee != nil {
		//fmt.Println(seek)
		//fmt.Println(version)
		//fmt.Println(len(buf))
		//fmt.Println(buf[seek: seek+50])
		//fmt.Println(ee)
		fmt.Println("Block not Find. Version:", version)
	}
	var mv, err = blk.ParseHead(buf, seek+1)
	return blk, mv, err
}
