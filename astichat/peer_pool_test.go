package astichat_test

import (
	"net"
	"testing"

	"github.com/asticode/go-astichat/astichat"
	"github.com/stretchr/testify/assert"
)

func TestPeerPool(t *testing.T) {
	var pp = astichat.NewPeerPool()
	var p1 = astichat.NewPeer(&net.UDPAddr{}, astichat.Chatterer{Username: "bob"})
	assert.Equal(t, 0, pp.Len())
	pp.Set(p1)
	assert.Equal(t, 1, pp.Len())
	var p2, ok = pp.Get("invalid")
	assert.False(t, ok)
	p2, ok = pp.Get("bob")
	assert.True(t, ok)
	assert.Equal(t, p2, p1)
	pp.Del("bob")
	assert.Equal(t, 0, pp.Len())
}
