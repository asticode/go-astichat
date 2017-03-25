package main

import (
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
	"github.com/rs/xlog"
)

// Server represents a server
type Server struct {
	channelQuit chan bool
	logger      xlog.Logger
	peerPool    *astichat.PeerPool
	server      *astiudp.Server
	startedAt   time.Time
}

// NewServer returns a new server
func NewServer(l xlog.Logger) *Server {
	l.Debug("Starting server")
	return &Server{
		channelQuit: make(chan bool),
		logger:      l,
		peerPool:    astichat.NewPeerPool(),
		server:      astiudp.NewServer(),
		startedAt:   time.Now(),
	}
}

// Init initialises the server
func (s *Server) Init(c Configuration) (o *Server, err error) {
	// Init server
	o = s
	o.server.Logger = o.logger
	if err = s.server.Init(c.ListenAddr); err != nil {
		return
	}

	// Set up listeners
	s.server.SetListener(astichat.EventNamePeerRegister, s.HandlePeerRegister())
	s.server.SetListener(astichat.EventNamePeerDisconnect, s.HandlePeerDisconnect())
	return
}

// Close closes the server
func (s *Server) Close() {
	s.server.Close()
	s.logger.Debug("Stopping server")
}

// HandleSignals handles signals
func (s *Server) HandleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func(s *Server) {
		for sig := range ch {
			s.logger.Debugf("Received signal %s", sig)
			s.Stop()
		}
	}(s)
}

// Stop stops the server
func (s *Server) Stop() {
	close(s.channelQuit)
}

// Wait is a blocking pattern
func (s *Server) Wait() {
	for {
		select {
		case <-s.channelQuit:
			return
		}
	}
}

// HandlePeerRegister handles the peer.register event
func (s *Server) HandlePeerRegister() astiudp.ListenerFunc {
	return func(as *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Peer is new to the pool
		var p *astichat.Peer
		var ok bool
		if p, ok = s.peerPool.Get(b.PublicKey); !ok {
			// Create peer
			p = astichat.NewPeer(addr, b.PublicKey)

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
func (s *Server) HandlePeerDisconnect() astiudp.ListenerFunc {
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
