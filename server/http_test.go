package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/asticode/go-astichat/astichat"
	main "github.com/asticode/go-astichat/server"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/xlog"
	"github.com/stretchr/testify/assert"
)

func TestHandleTokenPOST(t *testing.T) {
	// Init
	var l = xlog.NopLogger
	var s = astichat.NewMockedStorage()
	main.GenerateToken = func() string {
		return "new id"
	}
	main.Now = func() time.Time {
		return time.Unix(100, 0)
	}
	rw := httptest.NewRecorder()
	r := &http.Request{}
	r = r.WithContext(main.NewContextWithLogger(r.Context(), l))
	r = r.WithContext(main.NewContextWithStorage(r.Context(), s))

	// Empty username
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, rw.Code)

	// Username doesn't exist
	r.Form.Set("username", "bob")
	rw = httptest.NewRecorder()
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)

	// Username exists
	s.ChattererCreate("bob", &astichat.PublicKey{}, &astichat.PrivateKey{})
	rw = httptest.NewRecorder()
	main.HandleTokenPOST(rw, r, httprouter.Params{})
	assert.Equal(t, http.StatusOK, rw.Code)
	c, err := s.ChattererFetchByUsername("bob")
	assert.NoError(t, err)
	assert.Equal(t, "new id", c.Token)
	assert.Equal(t, time.Unix(100, 0), c.TokenAt)
}
