package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
)

// LDFlags
var (
	ClientPrivateKey string
	ServerHTTPAddr   string
	ServerPublicKey  string
	ServerUDPAddr    string
	Username         string
	Version          string
)

// TODO Use UI instead + think about go-mobile
// TODO One should be able to choose between client-2-server or client-2-client connections
// TODO Remove the configuration via flags and force using the UI
func main() {
	// Parse command
	var s = astiflag.Subcommand()
	flag.Parse()

	// Init configuration
	var c = NewConfiguration()

	// Init logger
	var l = astilog.New(c.Logger)

	// Create client
	var cl = NewClient(l)
	defer cl.Close()

	// Init client
	var err error
	if err = cl.Init(c); err != nil {
		l.Fatal(err)
	}

	// Handle signals
	cl.HandleSignals()

	// Switch on subcommand
	switch s {
	case "token":
		var token string
		if token, err = cl.Token(); err != nil {
			l.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, token)
	case "username":
		fmt.Fprintln(os.Stdout, cl.username)
	case "version":
		fmt.Fprintln(os.Stdout, cl.version)
	default:
		// Listen and read
		go cl.server.ListenAndRead()

		// Wait is the blocking pattern
		cl.Wait()
	}
}
