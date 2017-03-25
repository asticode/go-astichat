package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/pem"
	"io/ioutil"

	"crypto/rsa"
	"crypto/x509"

	"fmt"

	"encoding/base64"

	"github.com/rs/xlog"
)

// Client represents a client
type Client struct {
	channelQuit chan bool
	logger      xlog.Logger
	pemPath     string
	passphrase  string
	privateKey  *rsa.PrivateKey
	startedAt   time.Time
	version     string
}

// NewClient returns a new client
func NewClient(c Configuration, l xlog.Logger) *Client {
	l.Debug("Starting client")
	return &Client{
		channelQuit: make(chan bool),
		logger:      l,
		passphrase:  c.Passphrase,
		pemPath:     c.PEMPath,
		startedAt:   time.Now(),
		version:     Version,
	}
}

// Init initialises the client
func (c *Client) Init() (o *Client, err error) {
	// Retrieve pem data
	o = c
	var pemData []byte
	if len(c.pemPath) > 0 {
		if pemData, err = ioutil.ReadFile(c.pemPath); err != nil {
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

	// Decrypt block
	var b = block.Bytes
	if len(c.passphrase) > 0 {
		if b, err = x509.DecryptPEMBlock(block, []byte(c.passphrase)); err != nil {
			err = fmt.Errorf("Invalid passphrase: %s", err)
			return
		}
	}

	// Parse private key
	if o.privateKey, err = x509.ParsePKCS1PrivateKey(b); err != nil {
		return
	}
	return
}

// Close closes the client
func (c *Client) Close() {
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
