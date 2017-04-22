package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/asticode/go-astichat/astichat"
)

// sendHTTP sends a message to the HTTP server
func (c *Client) sendHTTP(method, pattern string, msg []byte) (o []byte, err error) {
	// Create new body
	var b astichat.Body
	if b, err = astichat.NewBody(msg, c.now.Time(), Username, c.serverPublicKey); err != nil {
		return
	}

	// Marshal
	var buf = &bytes.Buffer{}
	if err = json.NewEncoder(buf).Encode(b); err != nil {
		return
	}

	// Create request
	var req *http.Request
	if req, err = http.NewRequest(method, c.serverHTTPAddr+pattern, buf); err != nil {
		return
	}

	// Send request
	// TODO Add retry
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	// Unmarshal
	if err = json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return
	}

	// Process body
	if o, err = b.Process(c.now.Time(), c.privateKey); err != nil {
		return
	}

	return
}

// Now generates a new now
func (c *Client) Now() (now *astichat.Now, err error) {
	// Create request
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, c.serverHTTPAddr+"/now", nil); err != nil {
		return
	}

	// Send request
	var resp *http.Response
	if resp, err = c.httpClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	// Unmarshal
	var t time.Time
	if err = json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return
	}

	// Create now
	now = astichat.NewNow(t)
	return
}

// Token generates a new token
func (c *Client) Token() (o string, err error) {
	// Send
	var token astichat.Token
	if token, err = c.sendHTTP(http.MethodPost, "/token", astichat.MessageToken); err != nil {
		return
	}

	// Encode
	if o, err = token.Encode(c.serverPublicKey); err != nil {
		return
	}
	return
}
