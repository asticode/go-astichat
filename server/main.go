package main

import (
	"flag"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astimgo"
	"github.com/asticode/go-astitools/flag"
	"gopkg.in/mgo.v2"
)

func main() {
	// Parse command
	var s = astiflag.Subcommand()
	flag.Parse()

	// Init configuration
	var c = NewConfiguration()

	// Init logger
	var l = astilog.New(c.Logger)

	// Init builder
	var b = builder.New(c.Builder)
	b.Logger = l

	// Init mongo
	var ms *mgo.Session
	var err error
	if ms, err = astimgo.NewSession(c.Mongo); err != nil {
		l.Fatal(err)
	}
	defer ms.Close()

	// Init storage
	var stg = astichat.NewStorageMongo(l, ms)

	// Init server
	var srv *Server
	if srv, err = NewServer(l, c, b, stg).Init(c); err != nil {
		l.Fatal(err)
	}
	defer srv.Close()

	// Handle signals
	srv.HandleSignals()

	// Switch on subcommand
	switch s {
	default:
		// Listen and serve
		srv.ListenAndServe()

		// Wait is the blocking pattern
		srv.Wait()
	}
}
