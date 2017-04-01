package astichat

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	"encoding/pem"

	"crypto"
	"crypto/sha512"

	"gopkg.in/mgo.v2/bson"
)

// Vars
var (
	b64     = base64.StdEncoding
	keyBits = 4096
	pubHash = sha512.New()
	prvHash = crypto.SHA512
)

// EncryptedMessage represents an encrypted message
type EncryptedMessage struct {
	Hash      []byte `json:"hash,omitempty"`
	Message   []byte `json:"message,omitempty"`
	Signature []byte `json:"signature,omitempty"`
}

// NewEncryptedMessage encrypts a message
func NewEncryptedMessage(i []byte, pubDst *PublicKey, prvSrc *PrivateKey) (o EncryptedMessage, err error) {
	// Encrypt message with dst public key
	if o.Message, err = rsa.EncryptOAEP(pubHash, rand.Reader, pubDst.key, i, nil); err != nil {
		return
	}

	// Sign message with private key
	var pssh = prvHash.New()
	if _, err = pssh.Write(i); err != nil {
		return
	}
	o.Hash = pssh.Sum(nil)
	if o.Signature, err = rsa.SignPSS(rand.Reader, prvSrc.key, prvHash, o.Hash, nil); err != nil {
		return
	}
	return
}

// Decrypt decrypts a message
func (m EncryptedMessage) Decrypt(pubDst *PublicKey, prvSrc *PrivateKey) (o []byte, err error) {
	// Decrypt message with private key
	if o, err = rsa.DecryptOAEP(pubHash, rand.Reader, prvSrc.key, m.Message, nil); err != nil {
		return
	}

	// Check signature with public key
	if err = rsa.VerifyPSS(pubDst.key, prvHash, m.Hash, m.Signature, nil); err != nil {
		return
	}
	return
}

// PrivateKey represents a marshalable/unmarshalable private key
type PrivateKey struct {
	key        *rsa.PrivateKey
	passphrase string
	string     string
}

// NewPrivateKey generates a new private key
func NewPrivateKey(passphrase string) (p *PrivateKey, err error) {
	p = &PrivateKey{passphrase: passphrase}
	if p.key, err = rsa.GenerateKey(rand.Reader, keyBits); err != nil {
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

	// Base64 encode
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

// PublicKey represents a marshalable/unmarshalable public key
type PublicKey struct {
	key    *rsa.PublicKey
	string string
}

// NewPublicKey creates a new PublicKey based on a *rsa.PublicKey
func NewPublicKey(pub *rsa.PublicKey) *PublicKey {
	return &PublicKey{key: pub}
}

// Key returns the *rsa.PublicKey
func (p PublicKey) Key() *rsa.PublicKey {
	return p.key
}

// MarshalText allows PublicKey to implement the TextMarshaler interface
func (p PublicKey) MarshalText() (o []byte, err error) {
	var b []byte
	if b, err = x509.MarshalPKIXPublicKey(p.key); err != nil {
		return
	}
	o = make([]byte, b64.EncodedLen(len(b)))
	b64.Encode(o, b)
	p.string = string(o)
	return
}

// GetBSON implements bson.Getter.
func (p PublicKey) GetBSON() (interface{}, error) {
	return p.MarshalText()
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
	if p.key, ok = in.(*rsa.PublicKey); !ok {
		err = fmt.Errorf("Public key %s is not a *rsa.PublicKey", i)
	}
	p.string = string(i)
	return
}

// SetBSON implements bson.Setter.
func (p *PublicKey) SetBSON(raw bson.Raw) (err error) {
	var b []byte
	if err = raw.Unmarshal(&b); err != nil {
		return
	}
	return p.UnmarshalText(b)
}
