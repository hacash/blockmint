package p2p

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/hacash/blockmint/config"
)

var (
	globalInstanceP2PServer *P2PServer = nil
)

type P2PServer struct {
	config  p2p.Config
	running *p2p.Server
}

func GetGlobalInstanceP2PServer() *P2PServer {
	if globalInstanceP2PServer == nil {
		globalInstanceP2PServer = NewP2PService()
	}
	return globalInstanceP2PServer
}

func NewP2PService() *P2PServer {
	newser := &P2PServer{}
	key := NodeKey()
	key.Public()
	newser.config = p2p.Config{
		/////////////////////////////////////
		PrivateKey:      key,
		MaxPeers:        8,
		MaxPendingPeers: 4,
		DialRatio:       3,
		NodeDatabase:    config.GetCnfPathNodes(),
		ListenAddr:      ":3337",
		/////////////////////////////////////
	}
	s := &p2p.Server{Config: newser.config}
	newser.running = s
	return newser
}

func (this *P2PServer) Start() error {

	er := this.running.Start()
	if er != nil {
		panic(er)
	}

	return nil
}
