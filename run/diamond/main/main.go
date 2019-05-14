package main

import (
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/miner"
)

/**
 * 挖掘钻石矿工
 */
func main() {

	// config
	config.LoadConfigFile()

	blkminer := miner.GetGlobalInstanceHacashMiner()

	//
	startDiamondMiner(blkminer)

}

func startDiamondMiner(blkminer *miner.HacashMiner) {

	blkminer.GetPrevDiamondHash()

}
