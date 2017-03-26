package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	"github.com/rs/xlog"
)

// Server represents a server
type Server struct {
	channelQuit chan bool
	logger      xlog.Logger
	serverHTTP  *ServerHTTP
	serverUDP   *ServerUDP
	startedAt   time.Time
}

// NewServer returns a new server
func NewServer(l xlog.Logger, c Configuration, b *builder.Builder, stg astichat.Storage) *Server {
	l.Debug("Starting server")
	return &Server{
		channelQuit: make(chan bool),
		logger:      l,
		serverHTTP:  NewServerHTTP(l, c.Addr.HTTP, c.PathStatic, b, stg),
		serverUDP:   NewServerUDP(l, stg),
		startedAt:   time.Now(),
	}
}

// Init initialises the server
func (s *Server) Init(c Configuration) (o *Server, err error) {
	o = s

	// Init UDP server
	if err = o.serverUDP.Init(c); err != nil {
		return
	}

	// Init HTTP server
	if err = o.serverHTTP.Init(c); err != nil {
		return
	}
	return
}

// Close closes the server
func (s *Server) Close() {
	s.serverUDP.Close()
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

// ListenAndServe listens and serve
func (s *Server) ListenAndServe() {
	// UDP
	go s.serverUDP.ListenAndServe()

	// HTTP
	go s.serverHTTP.ListenAndServe()
}
