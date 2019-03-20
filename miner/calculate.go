package miner

import (
	"encoding/binary"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/blockmint/types/block"

	"github.com/hacash/x16rs"
)


func (this *HacashMiner) calculateNextBlock(newBlock block.Block, coinbase *transactions.Transaction_0_Coinbase) (block.Block) {

	// 读取多线程挖矿数量
	supercpu := config.Config.Miner.Supervene
	markwork := []byte(config.Config.Miner.Markword)
	minermsg := make([]byte, 16)
	copy(minermsg, markwork)
	if len(markwork) > 15 {
		panic("config.Config.Miner.Markword length too long over 15")
	}
	if supercpu == 0 || supercpu > 200 {
		panic("config.Config.Miner.Supervene value must in [0, 200]")
	}
	_, bigintdiff, _, _ := this.State.NextHeightTargetDifficultyCompact()
	targethashdiff := difficulty.BigToHash256(bigintdiff)

	//ttt, _ := hex.DecodeString("0000f00f27700000000000000000000000000000000000000000000000000000")
	//targethashdiff = ttt
	//fmt.Println(targethashdiff)
	//fmt.Println(hex.EncodeToString(targethashdiff))
	//panic("")

	// 启动并发线程
	var successch = make(chan uint32, supercpu)
	var successMsgi = uint8(0)
	var stopsign *byte = new(byte)
	*stopsign = 0
	for i:=uint8(0); i<uint8(supercpu); i++ {
		minermsg[15] = i
		coinbase.Message = fields.TrimString16(minermsg)
		// update mrkl root
		newBlock.SetMrklRoot(blocks.CalculateMrklRoot(newBlock.GetTransactions()))
		basestuff := blocks.CalculateBlockHashBaseStuff(newBlock)
		//fmt.Println(i, hex.EncodeToString(basestuff))
		go func(i uint8) {
			// 开始挖矿
			nonce_bytes := x16r.MinerNonceHashX16RS( stopsign, targethashdiff, basestuff )
			nonce := binary.BigEndian.Uint32(nonce_bytes)
			if nonce > 0 {
				// 成功挖出
				successMsgi = i
				successch <- nonce
			}
		}(i)
	}
	// 等待停止标记
	go func() {
		select {
		// 外部通知挖矿停止标记
		case stat := <-this.miningStatusCh:
			if stat == false {
				successch <- 0 // 停止其他所有挖矿
			}
		}
	}()

	// 等待挖矿成功
	success_nonce := <-successch
	*stopsign = 1 // 标记停止其他线程
	if success_nonce == 0 {
		// 挖矿退出
		return nil
	}

	// update mrkl root
	newBlock.SetNonce(success_nonce)
	minermsg[15] = successMsgi // 线程标记
	coinbase.Message = fields.TrimString16(minermsg)
	newBlock.SetMrklRoot(blocks.CalculateMrklRoot(newBlock.GetTransactions()))
	newBlock.Fresh() // 更新
	//hhh := newBlock.HashFresh()
	//fmt.Println("=========================")
	//fmt.Println(hhh)

	//panic("")
	return newBlock
}


func (this *HacashMiner) calculateTargetHash(newBlock block.Block) ([]byte, block.Block) {
	return nil, nil
}


func (this *HacashMiner) calculateTargetHashOneBlock(newBlockOne block.Block, stop chan bool) {

}
