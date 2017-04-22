package main

import (
	"encoding/json"
	"net"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
	"github.com/rs/xlog"
)

// ServerUDP represents an UDP server
// TODO Create rooms => creator controls who can join
// TODO Handle client-2-server connections with rabbitmq
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
	s.server.SetListener(astichat.EventNamePeerConnect, s.HandlePeerConnect())
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

// HandlePeerConnect handles the peer.connect event
func (s *ServerUDP) HandlePeerConnect() astiudp.ListenerFunc {
	return func(as *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Peer is new to the pool
		var p *astichat.Peer
		var ok bool
		if p, ok = s.peerPool.Get(b.Request.Username); !ok {
			// Retrieve chatterer
			var c astichat.Chatterer
			if c, err = s.storage.ChattererFetchByUsername(b.Request.Username); err != nil {
				return
			}

			// Process body
			var msg []byte
			if msg, err = b.Process(astichat.TimeNow(), c.ServerPrivateKey); err != nil {
				return
			}

			// Validate message
			if err = astichat.ValidateMessage(msg, astichat.MessageConnect); err != nil {
				return
			}

			// Create peer
			p = astichat.NewPeer(addr, c)

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
			// Peer is not the one which just connected
			if p.Username != pp.Username {
				// Marshal
				var msg []byte
				if msg, err = json.Marshal(p); err != nil {
					return
				}

				// Create new body
				if b, err = astichat.NewBody(msg, astichat.TimeNow(), "", pp.ClientPublicKey); err != nil {
					return
				}

				// Send peer.joined event
				s.logger.Debugf("Sending peer.joined to %s", pp)
				if err = as.Write(astichat.EventNamePeerJoined, b, pp.Addr); err != nil {
					return
				}
				ps = append(ps, pp)
			}
		}

		// Marshal
		var msg []byte
		if msg, err = json.Marshal(ps); err != nil {
			return
		}

		// Create new body
		if b, err = astichat.NewBody(msg, astichat.TimeNow(), "", p.ClientPublicKey); err != nil {
			return
		}

		// Send peer.connected event
		s.logger.Debugf("Sending peer.connected to %s", p)
		if err = as.Write(astichat.EventNamePeerConnected, b, p.Addr); err != nil {
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
		if p, ok := s.peerPool.Get(b.Request.Username); ok {
			// Process body
			var msg []byte
			if msg, err = b.Process(astichat.TimeNow(), p.Chatterer.ServerPrivateKey); err != nil {
				return
			}

			// Validate message
			if err = astichat.ValidateMessage(msg, astichat.MessageDisconnect); err != nil {
				return
			}

			// Delete from the pool
			s.peerPool.Del(p.Username)

			// Log
			s.logger.Infof("%s has left us", p)

			// Loop through peers
			for _, pp := range s.peerPool.Peers() {
				// Marshal
				var msg []byte
				if msg, err = json.Marshal(p); err != nil {
					return
				}

				// Create new body
				if b, err = astichat.NewBody(msg, astichat.TimeNow(), "", pp.ClientPublicKey); err != nil {
					return
				}

				// Send peer.disconnected event
				s.logger.Debugf("Sending peer.disconnected to %s", pp)
				if err = as.Write(astichat.EventNamePeerDisconnected, b, pp.Addr); err != nil {
					return
				}
			}
		}
		return
	}
}
