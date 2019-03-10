package p2p

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/core/coin"
	miner "github.com/hacash/blockmint/miner"
)

type handShakeStatusData struct {
	// network
	GenesisBlockHash []byte

	// version
	BlockVersion    uint8
	TransactionType uint8
	ActionKind      uint16

	// status
	CurrentBlockHeight uint64
	CurrentBlockHash   []byte
}

func CreateHandShakeStatusData() handShakeStatusData {
	blockminer := miner.GetGlobalInstanceHacashMiner()
	return handShakeStatusData{
		GenesisBlockHash:   coin.GetGenesisBlock().Hash(),
		BlockVersion:       1,
		TransactionType:    1,
		ActionKind:         0, // not use
		CurrentBlockHeight: blockminer.State.CurrentHeight(),
		CurrentBlockHash:   blockminer.State.CurrentBlockHash(),
	}
}

// 识别
func (this *handShakeStatusData) Confirm(other *handShakeStatusData) error {
	if bytes.Compare(this.GenesisBlockHash, other.GenesisBlockHash) != 0 {
		return fmt.Errorf("GenesisBlockHash is difference")
	}
	if this.BlockVersion != this.BlockVersion {
		return fmt.Errorf("BlockVersion is difference")
	}
	if this.TransactionType != this.TransactionType {
		return fmt.Errorf("TransactionType is difference")
	}
	return nil
}
