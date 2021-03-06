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
	RepairVersion   uint16

	// status
	CurrentBlockHeight uint64
	CurrentBlockHash   []byte
}

func CreateHandShakeStatusData() handShakeStatusData {
	blockminer := miner.GetGlobalInstanceHacashMiner()
	return handShakeStatusData{
		GenesisBlockHash:   coin.GetGenesisBlock().Hash(),
		BlockVersion:       1, // 不匹配的版本，不相连接
		TransactionType:    2, // 不匹配的版本，不相连接
		ActionKind:         6, // 不匹配的版本，不相连接
		RepairVersion:      2, // 不匹配的版本，不相连接
		CurrentBlockHeight: blockminer.State.CurrentHeight(),
		CurrentBlockHash:   blockminer.State.CurrentBlockHash(),
	}
}

// 识别
func (this *handShakeStatusData) Confirm(other *handShakeStatusData) error {
	if bytes.Compare(this.GenesisBlockHash, other.GenesisBlockHash) != 0 {
		return fmt.Errorf("GenesisBlockHash is difference")
	}
	if this.BlockVersion < other.BlockVersion ||
		this.TransactionType < other.TransactionType ||
		this.ActionKind < other.ActionKind {
		// 版本低于连接
		fmt.Println("[Error] Please update the miner software form https://hacash.org [/Error]")
	}

	if this.BlockVersion != other.BlockVersion {
		return fmt.Errorf("BlockVersion is difference")
	}
	if this.TransactionType != other.TransactionType {
		return fmt.Errorf("TransactionType is difference")
	}
	if this.ActionKind != other.ActionKind {
		return fmt.Errorf("ActionKind is difference")
	}
	if this.RepairVersion != other.RepairVersion {
		return fmt.Errorf("RepairVersion is difference")
	}
	return nil
}
