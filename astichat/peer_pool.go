package astichat

import "sync"

// PeerPool represents a pool of peers
type PeerPool struct {
	mutex *sync.Mutex
	pool  map[string]*Peer // The pool is indexed by public keys
}

// NewPeerPool creates a new peer pool
func NewPeerPool() *PeerPool {
	return &PeerPool{
		mutex: &sync.Mutex{},
		pool:  make(map[string]*Peer),
	}
}

// Del deletes a peer from the pool
func (pp *PeerPool) Del(publicKey *PublicKey) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	delete(pp.pool, publicKey.String())
}

// Get gets a peer from the pool
func (pp *PeerPool) Get(publicKey *PublicKey) (p *Peer, ok bool) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	p, ok = pp.pool[publicKey.String()]
	return
}

// Peers returns the peers in the pool
func (pp *PeerPool) Peers() (o []*Peer) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	for _, p := range pp.pool {
		o = append(o, p)
	}
	return
}

// Set sets a peer in the pool
func (pp *PeerPool) Set(p *Peer) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.pool[p.ClientPublicKey.String()] = p
}
