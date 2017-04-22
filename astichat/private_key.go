package astichat

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

// PrivateKey represents a marshalable/unmarshalable private key
type PrivateKey struct {
	key        *rsa.PrivateKey
	passphrase string
	string     string
}

// NewPrivateKey generates a new private key
func NewPrivateKey(passphrase string) (p *PrivateKey, err error) {
	p = &PrivateKey{passphrase: passphrase}
	if p.key, err = rsa.GenerateKey(rand.Reader, privateKeyBits); err != nil {
		return
	}
	return
}

// SetPassphrase sets the passphrase
func (p *PrivateKey) SetPassphrase(passphrase string) {
	p.passphrase = passphrase
}

// Key returns the *rsa.PrivateKey
func (p PrivateKey) Key() *rsa.PrivateKey {
	return p.key
}

// MarshalText allows PrivateKey to implement the TextMarshaler interface
func (p PrivateKey) MarshalText() (o []byte, err error) {
	// Convert it to pem
	var block = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(p.key),
	}

	// Encrypt the pem
	if len(p.passphrase) > 0 {
		if block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(p.passphrase), x509.PEMCipherAES256); err != nil {
			return
		}
	}

	// Encode to memory
	var b = pem.EncodeToMemory(block)

	// b64 encode
	o = make([]byte, b64.EncodedLen(len(b)))
	b64.Encode(o, b)
	p.string = string(o)
	return
}

// GetBSON implements bson.Getter.
func (p PrivateKey) GetBSON() (interface{}, error) {
	return p.MarshalText()
}

// String allows PrivateKey to implement the Stringer interface
func (p PrivateKey) String() string {
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

// UnmarshalText allows PrivateKey to implement the TextUnmarshaler interface
func (p *PrivateKey) UnmarshalText(i []byte) (err error) {
	// Base 64 decode
	var b = make([]byte, b64.DecodedLen(len(i)))
	var n int
	if n, err = b64.Decode(b, i); err != nil {
		return
	}
	b = b[:n]

	// Decode pem
	var block *pem.Block
	if block, _ = pem.Decode(b); block == nil {
		err = fmt.Errorf("No block found in pem %s", string(b))
		return
	}

	// Decrypt block
	b = block.Bytes
	if len(p.passphrase) > 0 {
		if b, err = x509.DecryptPEMBlock(block, []byte(p.passphrase)); err != nil {
			err = fmt.Errorf("Invalid passphrase: %s", err)
			return
		}
	}

	// Parse private key
	if p.key, err = x509.ParsePKCS1PrivateKey(b); err != nil {
		return
	}
	p.string = string(i)
	return
}

// SetBSON implements bson.Setter.
func (p *PrivateKey) SetBSON(raw bson.Raw) (err error) {
	var b []byte
	if err = raw.Unmarshal(&b); err != nil {
		return
	}
	return p.UnmarshalText(b)
}

// PublicKey returns the public part of the private key
func (p PrivateKey) PublicKey() (o *PublicKey, err error) {
	// Assert public key
	var pub *rsa.PublicKey
	var ok bool
	if pub, ok = p.key.Public().(*rsa.PublicKey); !ok {
		err = errors.New("Public key is not a *rsa.PublicKey")
		return
	}

	// Return
	o = NewPublicKey(pub)
	return
}
