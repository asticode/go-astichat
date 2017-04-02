package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
	"github.com/rs/xlog"
	"golang.org/x/crypto/ssh/terminal"
)

// Client represents a client
type Client struct {
	channelQuit     chan bool
	logger          xlog.Logger
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
func NewClient(l xlog.Logger) *Client {
	l.Debug("Starting client")
	return &Client{
		channelQuit: make(chan bool),
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
	cl.server.SetListener(astichat.EventNamePeerRegistered, cl.HandlePeerRegistered())
	cl.server.SetListener(astichat.EventNamePeerJoined, cl.HandlePeerJoined())
	cl.server.SetListener(astichat.EventNamePeerTyped, cl.HandlePeerTyped())

	// Resolve server addr
	cl.serverHTTPAddr = c.ServerHTTPAddr
	if cl.serverUDPAddr, err = net.ResolveUDPAddr("udp4", c.ServerUDPAddr); err != nil {
		return
	}

	// We're getting the hour from the server and incrementing it manually so that we don't have to trust local
	// time that could be modified by the user
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, cl.serverHTTPAddr+"/now", nil); err != nil {
		return
	}
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	var t time.Time
	if err = json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return
	}
	cl.now = astichat.NewNow(t)
	go cl.now.Update()

	// Get passphrase
	fmt.Println("Enter your passphrase:")
	var b []byte
	if b, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return
	}
	cl.privateKey.SetPassphrase(string(bytes.TrimSpace(b)))

	// Unmarshal client's private key
	if err = cl.privateKey.UnmarshalText([]byte(ClientPrivateKey)); err != nil {
		return
	}

	// Unmarshal server's public key
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

// HandleStart handles the start event
func (c *Client) HandleStart() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Encrypt message
		// TODO Encrypt token instead
		var m astichat.EncryptedMessage
		if m, err = astichat.NewEncryptedMessage(astichat.MessageRegister, c.serverPublicKey, c.privateKey); err != nil {
			return
		}

		// Write
		c.logger.Debugf("Sending peer.register to %s", c.serverUDPAddr)
		if err = s.Write(astichat.EventNamePeerRegister, astichat.Body{EncryptedMessage: m, Username: c.username}, c.serverUDPAddr); err != nil {
			return
		}
		return
	}
}

// Disconnect disconnects from the server
func (c *Client) Disconnect() (err error) {
	// Encrypt message
	// TODO Encrypt token instead
	var m astichat.EncryptedMessage
	if m, err = astichat.NewEncryptedMessage(astichat.MessageDisconnect, c.serverPublicKey, c.privateKey); err != nil {
		return
	}

	// Write
	c.logger.Debugf("Sending peer.disconnect to %s", c.serverUDPAddr)
	if err = c.server.Write(astichat.EventNamePeerDisconnect, astichat.Body{EncryptedMessage: m, Username: c.username}, c.serverUDPAddr); err != nil {
		return
	}
	return
}

// HandlePeerDisconnected handles the peer.disconnected event
func (c *Client) HandlePeerDisconnected() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var body astichat.Body
		if err = json.Unmarshal(payload, &body); err != nil {
			return
		}

		// Decrypt message
		var b []byte
		if b, err = body.EncryptedMessage.Decrypt(c.serverPublicKey, c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var p *astichat.Peer
		if err = json.Unmarshal(b, &p); err != nil {
			return
		}

		// Delete peer from pool
		c.peerPool.Del(p.Username)

		// Print
		fmt.Fprintf(os.Stdout, "%s has left\n", p)
		return
	}
}

// HandlePeerRegistered handles the peer.registered event
func (c *Client) HandlePeerRegistered() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var body astichat.Body
		if err = json.Unmarshal(payload, &body); err != nil {
			return
		}

		// Decrypt message
		var b []byte
		if b, err = body.EncryptedMessage.Decrypt(c.serverPublicKey, c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var ps []*astichat.Peer
		if err = json.Unmarshal(b, &ps); err != nil {
			return
		}

		// Print
		fmt.Fprintln(os.Stdout, "You're now connected")

		// Loop through peers
		for _, p := range ps {
			// Add peer to pool
			c.peerPool.Set(p)

			// Print
			fmt.Fprintf(os.Stdout, "%s is already here\n", p)
		}
		return
	}
}

// HandlePeerJoined handles the peer.joined event
func (c *Client) HandlePeerJoined() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var body astichat.Body
		if err = json.Unmarshal(payload, &body); err != nil {
			return
		}

		// Decrypt message
		var b []byte
		if b, err = body.EncryptedMessage.Decrypt(c.serverPublicKey, c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var p *astichat.Peer
		if err = json.Unmarshal(b, &p); err != nil {
			return
		}

		// Add peer to pool
		c.peerPool.Set(p)

		// Print
		fmt.Fprintf(os.Stdout, "%s has joined\n", p)
		return
	}
}

// Type captures typing and send it encrypted to all peers
func (c *Client) Type() {
	var s = bufio.NewScanner(bufio.NewReader(os.Stdin))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		// Execute the rest in a goroutine
		go func(line []byte) {
			// Loop through peers
			var err error
			for _, p := range c.peerPool.Peers() {
				// Encrypt message
				var m astichat.EncryptedMessage
				if m, err = astichat.NewEncryptedMessage(line, p.ClientPublicKey, c.privateKey); err != nil {
					c.logger.Errorf("%s while encrypting message %s to %s")
					continue
				}

				// Write message
				c.logger.Debugf("Sending peer.typed to %s", p)
				if err = c.server.Write(astichat.EventNamePeerTyped, astichat.Body{EncryptedMessage: m, Username: c.username}, p.Addr); err != nil {
					c.logger.Errorf("%s while sending peer.typed to %s", p)
					continue
				}
			}
		}(s.Bytes())
	}
}

// HandlePeerDisconnected handles the peer.disconnected event
func (c *Client) HandlePeerTyped() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b *astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Get peer from pool
		if p, ok := c.peerPool.Get(b.Username); ok {
			// Decrypt message
			var m []byte
			if m, err = b.EncryptedMessage.Decrypt(p.ClientPublicKey, c.privateKey); err != nil {
				return
			}

			// Print
			fmt.Fprintf(os.Stdout, "%s: %s\n", p, string(m))
		}
		return
	}
}
