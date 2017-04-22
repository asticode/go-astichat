package astichat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
)

// EncryptedMessage represents an encrypted message
type EncryptedMessage struct {
	IV      []byte `json:"iv,omitempty"`
	Key     []byte `json:"key,omitempty"`
	Message []byte `json:"message,omitempty"`
}

// NewEncryptedMessage encrypts a message
func NewEncryptedMessage(msg []byte, pubDst *PublicKey) (em EncryptedMessage, err error) {
	// Generate random key
	var key = make([]byte, aesKeyBits/8)
	if _, err = rand.Read(key); err != nil {
		return
	}

	// Create AES block
	var b cipher.Block
	if b, err = aes.NewCipher(key); err != nil {
		return
	}

	// Generate random IV
	em.IV = make([]byte, b.BlockSize())
	if _, err = rand.Read(em.IV); err != nil {
		return
	}

	// AES encrypt the message
	em.Message = make([]byte, len(msg))
	var cfb = cipher.NewCFBEncrypter(b, em.IV)
	cfb.XORKeyStream(em.Message, msg)

	// RSA encrypt the AES key
	if em.Key, err = rsa.EncryptOAEP(pubHash, rand.Reader, pubDst.key, key, nil); err != nil {
		return
	}
	return
}

// Decrypt decrypts a message
func (m EncryptedMessage) Decrypt(prvSrc *PrivateKey) (o []byte, err error) {
	// RSA decrypt the AES key
	var key []byte
	if key, err = rsa.DecryptOAEP(pubHash, rand.Reader, prvSrc.key, m.Key, nil); err != nil {
		return
	}

	// Create AES block
	var c cipher.Block
	if c, err = aes.NewCipher(key); err != nil {
		return
	}

	// AES decrypt the message
	o = make([]byte, len(m.Message))
	var cfb = cipher.NewCFBDecrypter(c, m.IV)
	cfb.XORKeyStream(o, m.Message)
	return
}
