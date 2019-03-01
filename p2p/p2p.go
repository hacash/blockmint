package p2p

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
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
		DialRatio:       4,
		NodeDatabase:    config.GetCnfPathNodes(),
		ListenAddr:      ":" + config.Config.P2p.Port.Node,
		Protocols:       []p2p.Protocol{MyProtocol()},
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

////////////////////////////////////////////

func MyProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    "MyProtocol",
		Version: 1,
		Length:  2,
		Run:     msgHandler,
	}
}

const messageId = 0
const messageId1 = 1

type Message string

func msgHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	fmt.Println("peer", peer.Name(), "connected.")
	p2p.SendItems(ws, messageId, "foo")
	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			fmt.Println("peer", peer.Name(), "disconnected")
			return err
		}
		// SendItems writes an RLP with the given code and data elements.
		// For a call such as:
		//
		//    SendItems(w, code, e1, e2, e3)
		//
		// the message payload will be an RLP list containing the items:
		//
		//    [e1, e2, e3]
		// 所以这里收消息应该定义为数组
		var myMessage [1]Message
		err = msg.Decode(&myMessage)
		if err != nil {
			// handle decode error
			continue
		}

		fmt.Println("code:", msg.Code, "receiver at:", msg.ReceivedAt, "msg:", myMessage)
		switch myMessage[0] {
		case "foo":
			err := p2p.SendItems(ws, messageId1, "bar")
			if err != nil {
				return err
			}
		default:
			fmt.Println("recv:", myMessage)
		}
	}
}
