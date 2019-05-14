package blocks

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/miner/x16rs"
	"github.com/hacash/blockmint/sys/err"
	typesblock "github.com/hacash/blockmint/types/block"

	"golang.org/x/crypto/sha3"
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
	if len(buf) < 1 {
		return nil, 0, fmt.Errorf("buf too short")
	}
	version := uint8(buf[seek])
	var blk, _ = NewBlockByVersion(version)
	var mv, err = blk.Parse(buf, seek+1)
	return blk, mv, err
}

func ParseBlockHead(buf []byte, seek uint32) (typesblock.Block, uint32, error) {
	version := uint8(buf[seek])
	var blk, ee = NewBlockByVersion(version)
	if ee != nil {
		fmt.Println("Block not Find. Version:", version)
	}
	var mv, err = blk.ParseHead(buf, seek+1)
	return blk, mv, err
}
func ParseExcludeTransactions(buf []byte, seek uint32) (typesblock.Block, uint32, error) {
	version := uint8(buf[seek])
	var blk, ee = NewBlockByVersion(version)
	if ee != nil {
		fmt.Println("Block not Find. Version:", version)
	}
	var mv, err = blk.ParseExcludeTransactions(buf, seek+1)
	return blk, mv, err
}

//////////////////////////////////

func CalculateBlockHash(block typesblock.Block) []byte {
	stuff := CalculateBlockHashBaseStuff(block)
	hashbase := sha3.Sum256(stuff)
	//fmt.Println( hex.EncodeToString( hashbase[:] ) )
	minerloopnum := int(block.GetHeight()/50000 + 1)
	if minerloopnum > 16 {
		minerloopnum = 16 // 8年时间上升到16次
	}
	return x16rs.HashX16RS_Optimize(minerloopnum, hashbase[:])
}

func CalculateBlockHashBaseStuff(block typesblock.Block) []byte {
	var buffer bytes.Buffer
	head, _ := block.SerializeHead()
	meta, _ := block.SerializeMeta()
	buffer.Write(head)
	buffer.Write(meta)
	return buffer.Bytes()
}

func CalculateMrklRoot(transactions []typesblock.Transaction) []byte {
	trslen := len(transactions)
	if trslen == 0 {
		return fields.EmptyZeroBytes32
	}
	hashs := make([][]byte, trslen)
	for i := 0; i < trslen; i++ {
		hashs[i] = transactions[i].Hash()
	}
	for true {
		if len(hashs) == 1 {
			return hashs[0]
		}
		hashs = hashMerge(hashs) // 两两归并
	}
	return nil
}

func hashMerge(hashs [][]byte) [][]byte {
	length := len(hashs)
	mgsize := length / 2
	if length%2 == 1 {
		mgsize = (length + 1) / 2
	}
	var mergehashs = make([][]byte, mgsize)
	for m := 0; m < length; m += 2 {
		var b1 bytes.Buffer
		b1.Write(hashs[m])
		h2 := hashs[m]
		if m+1 < length {
			h2 = hashs[m+1]
		}
		b1.Write(h2)
		digest := sha3.Sum256(b1.Bytes())
		mergehashs[m/2] = digest[:]
	}
	return mergehashs
}
