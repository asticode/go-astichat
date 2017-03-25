package astichat

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// Vars
var (
	b64 = base64.StdEncoding
)

// PublicKey represents a marshalable/unmarshalable public key
type PublicKey struct {
	*rsa.PublicKey
	string string
}

// MarshalText allows PublicKey to implement the TextMarshaler interface
func (p PublicKey) MarshalText() (o []byte, err error) {
	var b []byte
	if b, err = x509.MarshalPKIXPublicKey(p.PublicKey); err != nil {
		return
	}
	o = make([]byte, b64.EncodedLen(len(b)))
	b64.Encode(o, b)
	p.string = string(o)
	return
}

// String allows PublicKey to implement the Stringer interface
func (p PublicKey) String() string {
	if len(p.string) > 0 {
		return p.string
	}
	var b []byte
	var err error
	if b, err = p.MarshalText(); err != nil {
		return ""
	}
	return string(b)
}

// UnmarshalText allows PublicKey to implement the TextUnmarshaler interface
func (p *PublicKey) UnmarshalText(i []byte) (err error) {
	// Base 64 decode
	var b = make([]byte, b64.DecodedLen(len(i)))
	var n int
	if n, err = b64.Decode(b, i); err != nil {
		return
	}
	b = b[:n]

	// Parse
	var in interface{}
	if in, err = x509.ParsePKIXPublicKey(b); err != nil {
		return
	}

	// Assert
	var ok bool
	if p.PublicKey, ok = in.(*rsa.PublicKey); !ok {
		err = fmt.Errorf("Public key %s is not a *rsa.PublicKey", i)
	}
	p.string = string(i)
	return
}
