package astichat

import (
	"encoding/json"

	"bytes"
	"fmt"
	"time"

	"github.com/rs/xid"
)

// Token represents a token
type Token []byte

// GenerateToken allows testing functions using it
var GenerateToken = func() string {
	return xid.New().String()
}

// DecodeToken decodes a token
func DecodeToken(i string, prvSrc *PrivateKey) (t Token, err error) {
	// Base64 decode
	var b []byte
	if b, err = b64.DecodeString(i); err != nil {
		return
	}

	// Unmarshal
	var msg EncryptedMessage
	if err = json.Unmarshal(b, &msg); err != nil {
		return
	}

	// Decrypt message
	if t, err = msg.Decrypt(prvSrc); err != nil {
		return
	}
	return
}

// Encode encodes a token
func (t Token) Encode(pubDst *PublicKey) (o string, err error) {
	// Create encrypted message
	var msg EncryptedMessage
	if msg, err = NewEncryptedMessage(t, pubDst); err != nil {
		return
	}

	// Marshal
	var b []byte
	if b, err = json.Marshal(msg); err != nil {
		return
	}

	// Base64 encode
	o = b64.EncodeToString(b)
	return
}

// Validate validates a token against a chatterer
func (t Token) Validate(c Chatterer) error {
	// Validate content
	if !bytes.Equal(t, []byte(c.Token)) {
		return fmt.Errorf("Expected token %s but got %s", c.Token, t)
	}

	// Validate date
	var now = TimeNow()
	if c.TokenAt.After(now) || c.TokenAt.Before(now.Add(-5*time.Minute)) {
		return fmt.Errorf("Invalid token at %s since now is %s", c.TokenAt, now)
	}
	return nil
}
