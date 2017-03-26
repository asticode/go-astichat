package main

import (
	"encoding/json"
	"net"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
	"github.com/rs/xlog"
)

// ServerUDP represents an UDP server
type ServerUDP struct {
	logger   xlog.Logger
	peerPool *astichat.PeerPool
	server   *astiudp.Server
	storage  astichat.Storage
}

// NewServerUDP creates a new UDP sever
func NewServerUDP(l xlog.Logger, stg astichat.Storage) *ServerUDP {
	return &ServerUDP{
		logger:   l,
		peerPool: astichat.NewPeerPool(),
		server:   astiudp.NewServer(),
		storage:  stg,
	}
}

// Init initialises the UDP server
func (s *ServerUDP) Init(c Configuration) (err error) {
	// Init server
	s.server.Logger = s.logger
	if err = s.server.Init(c.Addr.UDP); err != nil {
		return
	}

	// Set up listeners
	s.server.SetListener(astichat.EventNamePeerRegister, s.HandlePeerRegister())
	s.server.SetListener(astichat.EventNamePeerDisconnect, s.HandlePeerDisconnect())
	return
}

// Close closes the UDP server
func (s *ServerUDP) Close() {
	s.server.Close()
}

// ListenAndServe listens and serve
func (s *ServerUDP) ListenAndServe() {
	s.server.ListenAndRead()
}

// HandlePeerRegister handles the peer.register event
func (s *ServerUDP) HandlePeerRegister() astiudp.ListenerFunc {
	return func(as *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Retrieve chatterer
		var c astichat.Chatterer
		if c, err = s.storage.ChattererFetchByPublicKey(b.PublicKey); err != nil {
			return
		}

		// Peer is new to the pool
		var p *astichat.Peer
		var ok bool
		if p, ok = s.peerPool.Get(c.PublicKey); !ok {
			// Create peer
			p = astichat.NewPeer(addr, c.PublicKey, c.Username)

			// Add peer to the pool
			s.peerPool.Set(p)

			// Log
			s.logger.Infof("Welcome to %s", p)
		} else {
			// TODO Peer is not new, check it has not updated its addr
		}

		// Loop through peers
		var ps []*astichat.Peer
		for _, pp := range s.peerPool.Peers() {
			// Peer is not the one which just registered
			if p.PublicKey.String() != pp.PublicKey.String() {
				// Send peer.joined event
				s.logger.Debugf("Sending peer.joined to %s", pp)
				if err = as.Write(astichat.EventNamePeerJoined, p, pp.Addr); err != nil {
					return
				}
				ps = append(ps, pp)
			}
		}

		// Send peer.registered event
		s.logger.Debugf("Sending peer.registered to %s", p)
		if err = as.Write(astichat.EventNamePeerRegistered, ps, p.Addr); err != nil {
			return
		}
		return
	}
}

// HandlePeerDisconnect handles the peer.disconnect event
func (s *ServerUDP) HandlePeerDisconnect() astiudp.ListenerFunc {
	return func(as *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Peer is in the pool
		if p, ok := s.peerPool.Get(b.PublicKey); ok {
			// Delete from the pool
			s.peerPool.Del(b.PublicKey)

			// Log
			s.logger.Infof("%s has left us", p)

			// Loop through peers
			for _, pp := range s.peerPool.Peers() {
				// Send peer.disconnected event
				s.logger.Debugf("Sending peer.disconnected to %s", pp)
				if err = as.Write(astichat.EventNamePeerDisconnected, p, pp.Addr); err != nil {
					return
				}
			}
		}
		return
	}
}
