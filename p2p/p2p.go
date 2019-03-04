package p2p

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/hacash/blockmint/config"
	"sync"
)

type P2PServer struct {
	config  p2p.Config
	running *p2p.Server
}

var (
	globalInstanceP2PServerMutex sync.Mutex
	globalInstanceP2PServer      *P2PServer = nil
)

func GetGlobalInstanceP2PServer() *P2PServer {
	globalInstanceP2PServerMutex.Lock()
	defer globalInstanceP2PServerMutex.Unlock()
	if globalInstanceP2PServer == nil {
		globalInstanceP2PServer = NewP2PService()
	}
	return globalInstanceP2PServer
}

func NewP2PService() *P2PServer {
	// 读取 main boot 节点
	urls := config.MainnetBootnodes
	var bootnodes = make([]*enode.Node, 0, len(urls))
	for _, url := range urls {
		node, err := enode.ParseV4(url)
		if err != nil {
			fmt.Println("Bootstrap URL invalid", "enode", url, "err", err)
		}
		bootnodes = append(bootnodes, node)
	}
	//
	protocolManager := GetGlobalInstanceProtocolManager()
	// 创建 server
	newser := &P2PServer{}
	key := NodeKey()
	newser.config = p2p.Config{
		StaticNodes: bootnodes,
		/////////////////////////////////////
		Name:            config.Config.P2p.Myname,
		PrivateKey:      key,
		MaxPeers:        16,
		MaxPendingPeers: 8,
		NodeDatabase:    config.GetCnfPathNodes(),
		ListenAddr:      ":" + config.Config.P2p.Port.Node,
		Protocols:       protocolManager.SubProtocols,
		NAT:             nat.Any(), // 支持内网穿透
		Logger:          log.New(),
		/////////////////////////////////////
	}
	s := &p2p.Server{Config: newser.config}
	newser.running = s
	return newser
}

func (this *P2PServer) Start() error {

	srv := this.running

	er := srv.Start()
	if er != nil {
		panic(er)
	}

	fmt.Printf("Hacash node started ... ...\n{ Name: %s, Url: %s }\n", srv.NodeInfo().Name, srv.NodeInfo().Enode)

	return nil
}
