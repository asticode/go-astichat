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
	var cl = NewClient(l)
	defer cl.Close()

	// Handle signals
	cl.HandleSignals()

	// Switch on subcommand
	var err error
	switch s {
	case "version":
		fmt.Fprintln(os.Stdout, cl.version)
	default:
		// Init client
		if err = cl.Init(c); err != nil {
			l.Fatal(err)
		}

		// Listen and read
		go cl.server.ListenAndRead()

		// Wait is the blocking pattern
		cl.Wait()
	}
}
