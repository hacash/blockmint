package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hacash/blockmint/config"
	"path"
)

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
		fmt.Errorf(fmt.Sprintf("Failed to generate node key: %v", err))
	}
	if err := crypto.SaveECDSA(keyfile, key); err != nil {
		fmt.Errorf(fmt.Sprintf("Failed to persist node key: %v", err))
	}
	return key
}
