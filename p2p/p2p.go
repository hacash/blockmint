package p2p

import (
	"fmt"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
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
	for _, url := range urls {
		node, err := enode.ParseV4(url)
		if err != nil {
			log.Error("bootstrap node url invalid", "enode", url, "err", err)
		}
		bootnodes = append(bootnodes, node)
	}
	//
	protocolManager := GetGlobalInstanceProtocolManager()
	// 创建 server
	newser := &P2PServer{
		Log: log,
	}
	key := NodeKey()
	newser.config = p2p.Config{
		BootstrapNodes: bootnodes,
		StaticNodes:    bootnodes,
		TrustedNodes:   bootnodes,
		/////////////////////////////////////
		Name:            config.Config.P2p.Myname,
		PrivateKey:      key,
		MaxPeers:        16,
		MaxPendingPeers: 8,
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
