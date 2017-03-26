package astichat

import "net"

// Peer represents a peer
type Peer struct {
	Addr *net.UDPAddr `json:"addr"`
	Chatterer
}

// NewPeer creates a new peer
func NewPeer(addr *net.UDPAddr, pk PublicKey, username string) *Peer {
	return &Peer{
		Addr: addr,
		Chatterer: Chatterer{
			PublicKey: pk,
			Username:  username,
		},
	}
}

// String allows Peer to implement the Stringer interface
func (p Peer) String() string {
	return p.Username
}
