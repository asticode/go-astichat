package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astiudp"
)

// HandleStart handles the start event
func (c *Client) HandleStart() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Create body
		var b astichat.Body
		if b, err = astichat.NewBody(astichat.MessageConnect, c.now.Time(), c.username, c.serverPublicKey); err != nil {
			return
		}

		// Write
		c.logger.Debugf("Sending peer.connect to %s", c.serverUDPAddr)
		if err = s.Write(astichat.EventNamePeerConnect, b, c.serverUDPAddr); err != nil {
			return
		}
		return
	}
}

// Disconnect disconnects from the server
func (c *Client) Disconnect() (err error) {
	if c.serverPublicKey != nil && c.privateKey != nil {
		// Create body
		var b astichat.Body
		if b, err = astichat.NewBody(astichat.MessageDisconnect, c.now.Time(), c.username, c.serverPublicKey); err != nil {
			return
		}

		// Write
		c.logger.Debugf("Sending peer.disconnect to %s", c.serverUDPAddr)
		if err = c.server.Write(astichat.EventNamePeerDisconnect, b, c.serverUDPAddr); err != nil {
			return
		}
	}
	return
}

// HandlePeerDisconnected handles the peer.disconnected event
func (c *Client) HandlePeerDisconnected() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Process body
		var msg []byte
		if msg, err = b.Process(c.now.Time(), c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var p *astichat.Peer
		if err = json.Unmarshal(msg, &p); err != nil {
			return
		}

		// Delete peer from pool
		c.peerPool.Del(p.Username)

		// Print
		fmt.Fprintf(os.Stdout, "%s has left\n", p)
		return
	}
}

// HandlePeerConnected handles the peer.connected event
func (c *Client) HandlePeerConnected() astiudp.ListenerFunc {
	return func(s *astiudp.Server, eventName string, payload json.RawMessage, addr *net.UDPAddr) (err error) {
		// Unmarshal
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Process body
		var msg []byte
		if msg, err = b.Process(c.now.Time(), c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var ps []*astichat.Peer
		if err = json.Unmarshal(msg, &ps); err != nil {
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
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Process body
		var msg []byte
		if msg, err = b.Process(c.now.Time(), c.privateKey); err != nil {
			return
		}

		// Unmarshal
		var p *astichat.Peer
		if err = json.Unmarshal(msg, &p); err != nil {
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
				// Create body
				var b astichat.Body
				if b, err = astichat.NewBody(line, c.now.Time(), c.username, p.ClientPublicKey); err != nil {
					return
				}

				// Write message
				c.logger.Debugf("Sending peer.typed to %s", p)
				if err = c.server.Write(astichat.EventNamePeerTyped, b, p.Addr); err != nil {
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
		var b astichat.Body
		if err = json.Unmarshal(payload, &b); err != nil {
			return
		}

		// Get peer from pool
		if p, ok := c.peerPool.Get(b.Request.Username); ok {
			// Process body
			var msg []byte
			if msg, err = b.Process(c.now.Time(), c.privateKey); err != nil {
				return
			}

			// Print
			fmt.Fprintf(os.Stdout, "%s: %s\n", p, string(msg))
		}
		return
	}
}
