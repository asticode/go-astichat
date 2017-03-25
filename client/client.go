package main

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
	"github.com/rs/xlog"
	"golang.org/x/crypto/ssh/terminal"
)

// Vars
var (
	pubHash = sha512.New()
	prvHash = crypto.SHA512
)

// Client represents a client
type Client struct {
	channelQuit chan bool
	logger      xlog.Logger
	peerPool    *astichat.PeerPool
	privateKey  *rsa.PrivateKey
	publicKey   astichat.PublicKey
	serverAddr  *net.UDPAddr
	server      *astiudp.Server
	startedAt   time.Time
	version     string
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
		version:     Version,
	}
}

// Init initialises the client
func (cl *Client) Init(c Configuration) (o *Client, err error) {
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
	if cl.serverAddr, err = net.ResolveUDPAddr("udp4", c.ServerAddr); err != nil {
		return
	}

	// Retrieve pem data
	o = cl
	var pemData []byte
	if len(c.PEMPath) > 0 {
		if pemData, err = ioutil.ReadFile(c.PEMPath); err != nil {
			return
		}
	} else {
		if pemData, err = base64.StdEncoding.DecodeString(PrivateKey); err != nil {
			return
		}
	}

	// Decode pem
	var block *pem.Block
	if block, _ = pem.Decode(pemData); block == nil {
		err = fmt.Errorf("No block found in pem %s", string(pemData))
		return
	}

	// Get passphrase
	fmt.Println("Enter your passphrase:")
	var b []byte
	if b, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return
	}
	var passphrase = string(bytes.TrimSpace(b))

	// Decrypt block
	b = block.Bytes
	if len(passphrase) > 0 {
		if b, err = x509.DecryptPEMBlock(block, []byte(passphrase)); err != nil {
			err = fmt.Errorf("Invalid passphrase: %s", err)
			return
		}
	}

	// Parse private key
	if o.privateKey, err = x509.ParsePKCS1PrivateKey(b); err != nil {
		return
	}

	// Parse public key
	var pk *rsa.PublicKey
	var ok bool
	if pk, ok = o.privateKey.Public().(*rsa.PublicKey); !ok {
		err = errors.New("Public key is not *rsa.PublicKey")
		return
	}
	o.publicKey = astichat.PublicKey{PublicKey: pk}

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
		// Write
		c.logger.Debugf("Sending peer.register to %s", c.serverAddr)
		if err = s.Write(astichat.EventNamePeerRegister, astichat.Body{PublicKey: c.publicKey}, c.serverAddr); err != nil {
			return
		}
		return
	}
}

// Disconnect disconnects from the server
func (c *Client) Disconnect() error {
	c.logger.Debugf("Sending peer.disconnect to %s", c.serverAddr)
	return c.server.Write(astichat.EventNamePeerDisconnect, astichat.Body{PublicKey: c.publicKey}, c.serverAddr)
}

// HandlePeerDisconnected handles the peer.disconnected event
func (c *Client) HandlePeerDisconnected() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var p *astichat.Peer
		if err = json.Unmarshal(payload, &p); err != nil {
			return
		}

		// Delete peer from pool
		c.peerPool.Del(p.PublicKey)

		// Print
		fmt.Fprintf(os.Stdout, "%s has left\n", p)
		return
	}
}

// HandlePeerRegistered handles the peer.registered event
func (c *Client) HandlePeerRegistered() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var ps []*astichat.Peer
		if err = json.Unmarshal(payload, &ps); err != nil {
			return
		}

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
		var p *astichat.Peer
		if err = json.Unmarshal(payload, &p); err != nil {
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
			for _, p := range c.peerPool.Peers() {
				// Encrypt message
				var message, hash, signature []byte
				var err error
				if message, hash, signature, err = c.encryptMessage(line, p.PublicKey.PublicKey); err != nil {
					c.logger.Errorf("%s while encrypting message %s to %s")
					continue
				}

				// Write message
				c.logger.Debugf("Sending peer.typed to %s", p)
				if err = c.server.Write(astichat.EventNamePeerTyped, astichat.BodyTyped{Body: astichat.Body{PublicKey: c.publicKey}, Hash: hash, Message: message, Signature: signature}, p.Addr); err != nil {
					c.logger.Errorf("%s while sending peer.typed to %s", p)
					continue
				}
			}
		}(s.Bytes())
	}
}

// encryptMessage encrypts a message
func (c *Client) encryptMessage(i []byte, pub *rsa.PublicKey) (o, hash, signature []byte, err error) {
	// Encrypt message with public key
	if o, err = rsa.EncryptOAEP(pubHash, rand.Reader, pub, i, nil); err != nil {
		return
	}

	// Sign message with private key
	var pssh = prvHash.New()
	if _, err = pssh.Write(i); err != nil {
		return
	}
	hash = pssh.Sum(nil)
	if signature, err = rsa.SignPSS(rand.Reader, c.privateKey, prvHash, hash, nil); err != nil {
		return
	}
	return
}

// decryptMessage decrypts a message
func (c *Client) decryptMessage(i, hash, signature []byte, pub *rsa.PublicKey) (o []byte, err error) {
	// Decrypt message with private key
	if o, err = rsa.DecryptOAEP(pubHash, rand.Reader, c.privateKey, i, nil); err != nil {
		return
	}

	// Check signature with public key
	if err = rsa.VerifyPSS(pub, prvHash, hash, signature, nil); err != nil {
		return
	}
	return
}

// HandlePeerDisconnected handles the peer.disconnected event
func (c *Client) HandlePeerTyped() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b *astichat.BodyTyped
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Get peer from pool
		if p, ok := c.peerPool.Get(b.PublicKey); ok {
			// Decrypt message
			var m []byte
			if m, err = c.decryptMessage(b.Message, b.Hash, b.Signature, p.PublicKey.PublicKey); err != nil {
				return
			}

			// Print
			fmt.Fprintf(os.Stdout, "%s: %s\n", p, string(m))
		}
		return
	}
}
