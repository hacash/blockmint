package miner

import (
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
)

type HacashMiner struct {
	CurrentBlock block.Block // 当前正在挖矿的区块

	TxPool service.TxPool
}

func NewHacashMiner() {
	return
}

// 开始挖矿
func (this *HacashMiner) Start() error {
	return nil
}

// 生成一个新区块
func (this *HacashMiner) CreateBlock() (block.Block, error) {
	return nil, nil
}
