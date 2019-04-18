package p2p

import (
	"fmt"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/sys/log"
	"sync"
)

type P2PServer struct {
	config  p2p.Config
	running *p2p.Server

	Log log.Logger
}

var (
	globalInstanceP2PServerMutex sync.Mutex
	globalInstanceP2PServer      *P2PServer = nil
)

func GetGlobalInstanceP2PServer() *P2PServer {
	globalInstanceP2PServerMutex.Lock()
	defer globalInstanceP2PServerMutex.Unlock()
	if globalInstanceP2PServer == nil {
		lg := config.GetGlobalInstanceLogger()
		globalInstanceP2PServer = NewP2PService(lg)
	}
	return globalInstanceP2PServer
}

func NewP2PService(log log.Logger) *P2PServer {
	// 读取 main boot 节点
	urls := config.MainnetBootnodes
	var bootnodes = make([]*enode.Node, 0, len(urls))
	var bootnodesV5 = make([]*discv5.Node, 0, len(urls))
	for _, url := range urls {
		node, err1 := enode.ParseV4(url)
		nodeV5, err2 := discv5.ParseNode(url)
		if err1 != nil || err2 != nil {
			log.Error("bootstrap node url invalid", "enode", url, "err1", err1, "err2", err2)
		}
		bootnodes = append(bootnodes, node)
		bootnodesV5 = append(bootnodesV5, nodeV5)
	}
	//
	protocolManager := GetGlobalInstanceProtocolManager()
	// 创建 server
	newser := &P2PServer{
		Log: log,
	}
	maxpeernum := int(config.Config.P2p.Maxpeernum)
	if maxpeernum == 0 {
		maxpeernum = 16
	}
	// fmt.Println(maxpeernum)
	key := NodeKey()
	newser.config = p2p.Config{
		BootstrapNodes: bootnodes,
		StaticNodes:    bootnodes,
		TrustedNodes:   bootnodes,
		// NoDiscovery:      true,
		DiscoveryV5:      true,
		BootstrapNodesV5: bootnodesV5,
		/////////////////////////////////////
		Name:            config.Config.P2p.Myname,
		PrivateKey:      key,
		MaxPeers:        maxpeernum,
		MaxPendingPeers: 100,
		NodeDatabase:    config.GetCnfPathNodes(),
		ListenAddr:      ":" + config.Config.P2p.Port.Node,
		Protocols:       protocolManager.SubProtocols,
		NAT:             nat.Any(), // 支持内网穿透
		Logger:          ethlog.New(),
		/////////////////////////////////////
	}
	s := &p2p.Server{Config: newser.config}
	newser.running = s
	return newser
}

func (this *P2PServer) GetServer() *p2p.Server {
	return this.running
}

func (this *P2PServer) Start() error {

	srv := this.running

	er := srv.Start()
	if er != nil {
		panic(er)
	}
	this.Log.Note(fmt.Sprintf("p2p node started name %s url %s", srv.NodeInfo().Name, srv.NodeInfo().Enode))

	return nil
}
