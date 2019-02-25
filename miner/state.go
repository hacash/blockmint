package miner

import (
	"bytes"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	"os"
	"path"
)

var distFileSize = block1def.ByteSizeBlockBeforeTransaction

type MinerState struct {
	prevBlockHead block.Block
}

func NewMinerState() *MinerState {
	return &MinerState{}
}

func (this *MinerState) getDistFile() *os.File {
	dir := config.GetCnfPathMinerState()
	filename := path.Join(dir, "state.dat")
	file.CreatePath(dir)
	file, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	return file
}

func (this *MinerState) FlushSave() {
	head := new(bytes.Buffer)
	b1, _ := this.prevBlockHead.SerializeHead()
	b2, _ := this.prevBlockHead.SerializeMeta()
	head.Write(b1)
	head.Write(b2)
	//
	file := this.getDistFile()

	//fmt.Println( "miner state: " + hex.EncodeToString(head.Bytes()) )
	file.Write(head.Bytes())
	file.Close()
}

func (this *MinerState) DistLoad() {
	file := this.getDistFile()
	valuebytes := make([]byte, distFileSize)
	rdlen, e := file.Read(valuebytes)
	if e != nil && rdlen == distFileSize {
		this.prevBlockHead = blocks.NewEmptyBlock_v1(nil)
		seek, _ := this.prevBlockHead.ParseHead(valuebytes, 0)
		seek, _ = this.prevBlockHead.ParseMeta(valuebytes, seek)
	}
}
