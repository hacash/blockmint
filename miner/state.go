package miner

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"math/big"
	"os"
	"path"
)

var (                             // 4294967295
	LowestDifficultyCompact = uint32(4294967294) // 首次调整难度前的预设难度值

	// 保存文件尺寸
	distFileSize = block1def.ByteSizeBlockBeforeTransaction + 8
)

type MinerState struct {
	prevBlockHead         block.Block
	prev288BlockTimestamp uint64 // 上一个288倍数区块的创建时间
	Log                   log.Logger
}

func NewMinerState(log log.Logger) *MinerState {
	return &MinerState{
		Log: log,
	}
}

// 获取
func (this *MinerState) GetBlockHead() block.Block {
	return this.prevBlockHead
}

// 修改矿工状态
func (this *MinerState) SetNewBlock(block block.Block) {
	this.prevBlockHead = block
	if block.GetHeight()%uint64(config.ChangeDifficultyBlockNumber) == 0 {
		this.prev288BlockTimestamp = block.GetTimestamp()
	}
	this.FlushSave()
}

// 获取下一个区块的难度值
func (this *MinerState) TargetDifficultyCompact(height uint64, print *string) (*big.Int, uint32) {
	// 预设难度
	if height < config.ChangeDifficultyBlockNumber {
		return difficulty.Uint32ToBig(LowestDifficultyCompact), LowestDifficultyCompact
	}
	head := this.prevBlockHead
	//targetdiff := difficulty.CalculateNextWorkTarget(
	targetbig, targetdiff := difficulty.CalculateNextTargetDifficulty(
		head.GetDifficulty(),
		height,
		this.prev288BlockTimestamp,
		head.GetTimestamp(),
		config.EachBlockTakesTime,
		config.ChangeDifficultyBlockNumber,
		print,
	)
	//fmt.Println("targetdiff", targetdiff)
	return targetbig, targetdiff
}

// 获取下一个区块的难度值
func (this *MinerState) NextHeightTargetDifficultyCompact() (uint64, uint32, *string) {
	nextHei := this.CurrentHeight() + 1
	print := new(string)
	_, target := this.TargetDifficultyCompact(nextHei, print)
	return nextHei, target, print
}

func (this *MinerState) CurrentHeight() uint64 {
	if this.prevBlockHead == nil {
		return 0
	}
	return this.prevBlockHead.GetHeight()
}
func (this *MinerState) CurrentBlockHash() []byte {
	if this.prevBlockHead == nil {
		return coin.GetGenesisBlock().Hash()
	}
	return this.prevBlockHead.Hash()
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
	b3 := make([]byte, 8)
	binary.BigEndian.PutUint64(b3, this.prev288BlockTimestamp)
	head.Write(b1)
	head.Write(b2)
	head.Write(b3)
	//
	file := this.getDistFile()

	//fmt.Println( "miner state save: " + hex.EncodeToString(head.Bytes()) )
	file.WriteAt(head.Bytes(), 0)
	file.Close()
}

func (this *MinerState) FetchLoad() {
	file := this.getDistFile()
	defer file.Close()
	valuebytes := make([]byte, distFileSize)
	rdlen, e := file.Read(valuebytes)
	if e == nil && rdlen >= distFileSize {
		this.prevBlockHead = blocks.NewEmptyBlock_v1(nil)
		seek, _ := this.prevBlockHead.ParseHead(valuebytes, 1)
		seek, _ = this.prevBlockHead.ParseMeta(valuebytes, seek)
		this.prev288BlockTimestamp = binary.BigEndian.Uint64(valuebytes[seek : seek+8])
		head := this.prevBlockHead
		this.Log.Note("miner state load from file", "height", head.GetHeight(), "hash", hex.EncodeToString(head.Hash()), "difficulty", head.GetDifficulty())
	} else {
		this.Log.Note("no find miner state file, set state with genesis block")
		genesis := coin.GetGenesisBlock()
		this.prevBlockHead = genesis                        // 创世
		this.prev288BlockTimestamp = genesis.GetTimestamp() // uint64(time.Now().Unix())
	}
}
