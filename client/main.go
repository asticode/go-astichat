package main

import (
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
)

// LDFlags
var (
	PrivateKey string
	ServerAddr string
	Version    string
)

func main() {
	// Parse command
	var s = astiflag.Subcommand()
	flag.Parse()

	// Init configuration
	var c = NewConfiguration()

	// Init logger
	var l = astilog.New(c.Logger)

	// Init client
	var cl *Client
	var err error
	if cl, err = NewClient(l).Init(c); err != nil {
		l.Fatal(err)
	}
	defer cl.Close()

	// Handle signals
	cl.HandleSignals()

	// Switch on subcommand
	switch s {
	case "version":
		l.Infof("Version is %s", cl.version)
	default:
		// Listen and read
		go cl.server.ListenAndRead()

		// Wait is the blocking pattern
		cl.Wait()
	}
}
