package astichat

import (
	"fmt"
	"net"
)

// Peer represents a peer
type Peer struct {
	Addr *net.UDPAddr `json:"addr"`
	Chatterer
}

// NewPeer creates a new peer
func NewPeer(addr *net.UDPAddr, c Chatterer) *Peer {
	return &Peer{
		Addr:      addr,
		Chatterer: c,
	}
}

// String allows Peer to implement the Stringer interface
func (p Peer) String() string {
	return fmt.Sprintf("%s@%s", p.Username, p.Addr)
}
