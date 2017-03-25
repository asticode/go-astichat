package astichat

import "net"

// Peer represents a peer
type Peer struct {
	Addr      *net.UDPAddr `json:"addr"`
	PublicKey PublicKey    `json:"public_key"`
}

// NewPeer creates a new peer
func NewPeer(addr *net.UDPAddr, pk PublicKey) *Peer {
	return &Peer{
		Addr:      addr,
		PublicKey: pk,
	}
}

// String allows Peer to implement the Stringer interface
// TODO Use local mapping between public key <--> username
func (p Peer) String() string {
	return p.Addr.String()
}
