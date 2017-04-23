package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astiudp"
	"golang.org/x/crypto/ssh/terminal"
)

// Client represents a client
type Client struct {
	channelQuit     chan bool
	httpClient      *http.Client
	logger          astilog.Logger
	now             *astichat.Now
	peerPool        *astichat.PeerPool
	privateKey      *astichat.PrivateKey
	server          *astiudp.Server
	serverHTTPAddr  string
	serverPublicKey *astichat.PublicKey
	serverUDPAddr   *net.UDPAddr
	startedAt       time.Time
	username        string
	version         string
}

// NewClient returns a new client
func NewClient(l astilog.Logger) *Client {
	l.Debug("Starting client")
	return &Client{
		channelQuit: make(chan bool),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		logger:      l,
		peerPool:    astichat.NewPeerPool(),
		server:      astiudp.NewServer(),
		startedAt:   time.Now(),
		username:    Username,
		version:     Version,
	}
}

// Init initialises the client
func (cl *Client) Init(c Configuration) (err error) {
	// Init server
	cl.server.Logger = cl.logger
	if err = cl.server.Init(c.ListenAddr); err != nil {
		return
	}

	// Set up server listeners
	cl.server.SetListener(astiudp.EventNameStart, cl.HandleStart())
	cl.server.SetListener(astichat.EventNamePeerDisconnected, cl.HandlePeerDisconnected())
	cl.server.SetListener(astichat.EventNamePeerConnected, cl.HandlePeerConnected())
	cl.server.SetListener(astichat.EventNamePeerJoined, cl.HandlePeerJoined())
	cl.server.SetListener(astichat.EventNamePeerTyped, cl.HandlePeerTyped())

	// Resolve server addr
	cl.serverHTTPAddr = ServerHTTPAddr
	if cl.serverUDPAddr, err = net.ResolveUDPAddr("udp4", ServerUDPAddr); err != nil {
		return
	}

	// We're getting the hour from the server and incrementing it manually so that we don't have to trust local
	// time that could be modified by the user
	if cl.now, err = cl.Now(); err != nil {
		return
	}

	// Get passphrase
	fmt.Println("Enter your passphrase:")
	var b []byte
	if b, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return
	}
	cl.privateKey = &astichat.PrivateKey{}
	cl.privateKey.SetPassphrase(string(bytes.TrimSpace(b)))

	// Unmarshal client's private key
	if err = cl.privateKey.UnmarshalText([]byte(ClientPrivateKey)); err != nil {
		return
	}

	// Unmarshal server's public key
	cl.serverPublicKey = &astichat.PublicKey{}
	if err = cl.serverPublicKey.UnmarshalText([]byte(ServerPublicKey)); err != nil {
		return
	}

	// Init Typing
	go cl.Type()
	return
}

// Close closes the client
func (c *Client) Close() {
	c.Disconnect()
	c.server.Close()
	c.logger.Debug("Stopping client")
}

// HandleSignals handles signals
func (c *Client) HandleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func(c *Client) {
		for s := range ch {
			c.logger.Debugf("Received signal %s", s)
			c.Stop()
		}
	}(c)
}

// Stop stops the client
func (c *Client) Stop() {
	close(c.channelQuit)
}

// Wait is a blocking pattern
func (c *Client) Wait() {
	for {
		select {
		case <-c.channelQuit:
			return
		}
	}
}
