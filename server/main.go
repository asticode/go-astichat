package main

import (
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
)

func main() {
	// Parse command
	var s = astiflag.Subcommand()
	flag.Parse()

	// Init configuration
	var c = NewConfiguration()

	// Init logger
	var l = astilog.New(c.Logger)

	// Init server
	var srv *Server
	var err error
	if srv, err = NewServer(l).Init(c); err != nil {
		l.Fatal(err)
	}
	defer srv.Close()

	// Handle signals
	srv.HandleSignals()

	// Switch on subcommand
	switch s {
	default:
		// Listen and read
		go srv.server.ListenAndRead()

		// Wait is the blocking pattern
		srv.Wait()
	}
}
