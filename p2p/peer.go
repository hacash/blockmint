package p2p

import (
	"fmt"
	"github.com/ethereum/go-ethereum/p2p"
	"sync"
)

// peerSet represents the collection of active peers currently participating in
// the Ethereum sub-protocol.
type peerSet struct {
	peers  map[string]*peer
	lock   sync.RWMutex
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

///////

type peer struct {
	id string

	*p2p.Peer
	rw p2p.MsgReadWriter
}

func newPeer(p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return &peer{
		Peer: p,
		rw:   rw,
		id:   fmt.Sprintf("%x", p.ID().Bytes()[:8]),
	}
}
