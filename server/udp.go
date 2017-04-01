package main

import (
	"encoding/json"
	"net"

	"fmt"

	"bytes"

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
		var body astichat.Body
		if err = json.Unmarshal(payload, &body); err != nil {
			return
		}

		// Retrieve chatterer
		var c astichat.Chatterer
		if c, err = s.storage.ChattererFetchByPublicKey(body.PublicKey); err != nil {
			return
		}

		// Peer is new to the pool
		var p *astichat.Peer
		var ok bool
		if p, ok = s.peerPool.Get(c.ClientPublicKey); !ok {
			// Decrypt message
			var b []byte
			if b, err = body.EncryptedMessage.Decrypt(c.ClientPublicKey, c.ServerPrivateKey); err != nil {
				return
			}

			// Validate message
			if !bytes.Equal(b, astichat.MessageRegister) {
				err = fmt.Errorf("Invalid message register %s", string(b))
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
			// Peer is not the one which just registered
			if p.ClientPublicKey.String() != pp.ClientPublicKey.String() {
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
		var body astichat.Body
		if err = json.Unmarshal(payload, &body); err != nil {
			return
		}

		// Peer is in the pool
		if p, ok := s.peerPool.Get(body.PublicKey); ok {
			// Decrypt message
			var b []byte
			if b, err = body.EncryptedMessage.Decrypt(p.ClientPublicKey, p.ServerPrivateKey); err != nil {
				return
			}

			// Validate message
			if !bytes.Equal(b, astichat.MessageDisconnect) {
				err = fmt.Errorf("Invalid message disconnect %s", string(b))
				return
			}

			// Delete from the pool
			s.peerPool.Del(p.ClientPublicKey)

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
