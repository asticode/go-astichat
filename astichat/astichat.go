package astichat

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

// Vars
var (
	aesKeyBits     = 256
	b64            = base64.StdEncoding
	privateKeyBits = 4096
	pubHash        = sha512.New()
)

// Event names
const (
	EventNamePeerConnect      = "peer.connect"
	EventNamePeerConnected    = "peer.connected"
	EventNamePeerDisconnect   = "peer.disconnect"
	EventNamePeerDisconnected = "peer.disconnected"
	EventNamePeerJoined       = "peer.joined"
	EventNamePeerTyped        = "peer.typed"
)

// Messages
var (
	MessageConnect    = []byte("connect")
	MessageDisconnect = []byte("disconnect")
	MessageToken      = []byte("token")
)

// ValidateMessage validates a message
func ValidateMessage(msg, expected []byte) (err error) {
	if !bytes.Equal(msg, expected) {
		err = fmt.Errorf("Expected message %s but got %s", expected, msg)
	}
	return
}
