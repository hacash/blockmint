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

// eth protocol message codes
const (
	// Protocol messages belonging to eth/62
	StatusMsg           = 0x00
	TxMsg               = 0x01
	GetSyncBlocksMsg    = 0x02
	SyncBlocksMsg       = 0x03
	NewBlockExcavateMsg = 0x04 // 新区块被挖出
	HeightHigherMsg     = 0x05 // 发现新区块高度

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
