package miner

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	"os"
	"path"
	"sync"
)

var distFileSize = block1def.ByteSizeBlockBeforeTransaction + 8

type MinerState struct {
	prevBlockHead         block.Block
	prev288BlockTimestamp uint64 // 上一个288倍数区块的创建时间

	lock sync.Mutex
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
	this.lock.Lock()
	defer this.lock.Unlock()

	head := new(bytes.Buffer)
	b1, _ := this.prevBlockHead.SerializeHead()
	b2, _ := this.prevBlockHead.SerializeMeta()
	b3 := make([]byte, 8)
	binary.BigEndian.PutUint64(b3, this.prev288BlockTimestamp)
	head.Write(b1)
	head.Write(b2)
	head.Write(b3)
	//
	file := this.getDistFile()

	//fmt.Println( "miner state save: " + hex.EncodeToString(head.Bytes()) )
	file.Write(head.Bytes())
	file.Close()
}

func (this *MinerState) DistLoad() {
	this.lock.Lock()
	defer this.lock.Unlock()

	file := this.getDistFile()
	valuebytes := make([]byte, distFileSize)
	rdlen, e := file.Read(valuebytes)
	if e == nil && rdlen >= distFileSize {
		this.prevBlockHead = blocks.NewEmptyBlock_v1(nil)
		seek, _ := this.prevBlockHead.ParseHead(valuebytes, 1)
		seek, _ = this.prevBlockHead.ParseMeta(valuebytes, seek)
		this.prev288BlockTimestamp = binary.BigEndian.Uint64(valuebytes[seek : seek+8])
	}
	file.Close()
	//fmt.Println("123")
	return
}
