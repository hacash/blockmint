package p2p

import (
	"github.com/ethereum/go-ethereum/p2p"
)

const (

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
)

type ProtocolManager struct {
	maxPeers int // 最大节点数量

	txsCh chan []byte // 交易广播

	peers *peerSet
}

func NewProtocolManager() (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		peers: newPeerSet(),
	}
	return manager, nil
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan []byte, txChanSize)
	go pm.txBroadcastLoop()

}

func (pm *ProtocolManager) newPeer(p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(p, rw)
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case tx := <-pm.txsCh:
			pm.BroadcastTx(tx)

			// Err() channel will be closed when unsubscribing.
			//case <-pm.txsSub.Err():
			//	return
		}
	}
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTx(tx []byte) {

	//peer.AsyncSendTransactions(tx)
}
