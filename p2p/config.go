package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/types/block"
	"os"
	"path"
)

const ProtocolMaxMsgSize = 2 * 1024 * 1024 // Maximum cap on the size of a protocol message

const SyncBlockDataMaxBlockNumber = 500 // 一次读取区块最大的数量上线

// protocol message codes
const (
	// Protocol messages
	StatusMsg           = 0x00
	TxMsg               = 0x01
	GetSyncBlocksMsg    = 0x02
	SyncBlocksMsg       = 0x03
	NewBlockExcavateMsg = 0x04 // 新区块被挖出
	HeightHigherMsg     = 0x05 // 发现新区块高度
	GetSyncHashsMsg     = 0x06
	SyncHashsMsg        = 0x07

	// p2p 状态
	SimpleStatus     = 0x00 // 普通状态
	SyncHashsStatus  = 0x01 // 正在检查分叉状态，是否需要回退
	SyncBlocksStatus = 0x02 // 正在同步区块
	SyncTxsStatus    = 0x03 // 正在同步交易池

)

type MsgDataGetSyncBlocks struct {
	StartHeight uint64
}

type MsgDataSyncBlocks struct {
	FromHeight uint64
	ToHeight   uint64
	Datas      string
}

type MsgDataNewBlock struct {
	block  block.Block
	Height uint64
	Datas  string
}

type MsgDataHeightHigher struct {
	Height uint64
}

type MsgDataGetSyncHashs struct {
	StartHeight uint64
	EndHeight   uint64
}

type MsgDataSyncHashs struct {
	StartHeight uint64
	EndHeight   uint64
	Hashs       []string
}

// NodeKey retrieves the currently configured private key of the node, checking
// first any manually set key, falling back to the one found in the configured
// data folder. If no key can be found, a new one is generated.
func NodeKey() *ecdsa.PrivateKey {

	datadirPrivateKey := config.GetCnfPathNodes()

	keyfile := path.Join(datadirPrivateKey, "node.private.key")

	if key, err := crypto.LoadECDSA(keyfile); err == nil {
		return key
	}
	// No persistent key found, generate and store a new one.
	key, err := crypto.GenerateKey()
	if err != nil {
		fmt.Errorf(fmt.Sprintf("Failed to generate node key: %v \n", err))
	}
	if err := os.MkdirAll(datadirPrivateKey, 0700); err != nil {
		fmt.Errorf(fmt.Sprintf("Failed to persist node key: %v \n", err))
		return key
	}
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		fmt.Errorf(fmt.Sprintf("Failed to persist node key: %v \n", err))
	}
	return key
}
