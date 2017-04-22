package astichat

import (
	"errors"
	"fmt"
	"time"
)

// Body represents a body
type Body struct {
	Error   *BodyError   `json:"error,omitempty"`
	Request *BodyRequest `json:"request,omitempty"`
}

// BodyError represents an error body
type BodyError struct {
	Message string `json:"message,omitempty"`
}

// BodyRequest represents a request body
type BodyRequest struct {
	CreatedAt time.Time        `json:"created_at,omitempty"`
	Message   EncryptedMessage `json:"message,omitempty"`
	Username  string           `json:"username,omitempty"`
}

// TimeNow allows testing functions using it
var TimeNow = func() time.Time {
	return time.Now()
}

// NewBody creates a new body
func NewBody(msg []byte, now time.Time, username string, pubDst *PublicKey) (b Body, err error) {
	// Init
	b = Body{Request: &BodyRequest{CreatedAt: now, Username: username}}

	// Encrypt
	if b.Request.Message, err = NewEncryptedMessage(msg, pubDst); err != nil {
		return
	}
	return
}

// Decrypt decrypts the body
func (b Body) Process(now time.Time, prvSrc *PrivateKey) (msg []byte, err error) {
	// Check error
	if b.Error != nil {
		err = errors.New(b.Error.Message)
		return
	}

	// Validate the request's creation date
	if b.Request.CreatedAt.After(now.Add(5*time.Second)) || b.Request.CreatedAt.Before(now.Add(-5*time.Second)) {
		err = fmt.Errorf("Request creation date %s is invalid compared to now %s", b.Request.CreatedAt, now)
		return
	}

	// Decrypt the request
	if msg, err = b.Request.Message.Decrypt(prvSrc); err != nil {
		return
	}
	return
}
