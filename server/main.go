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
	astilog.SetLogger(astilog.New(c.Logger))

	// Init builder
	var b = builder.New(c.Builder)

	// Init mongo
	var ms *mgo.Session
	var err error
	if ms, err = astimgo.NewSession(c.Mongo); err != nil {
		astilog.Fatal(err)
	}
	defer ms.Close()

	// Init storage
	var stg = astichat.NewStorageMongo(ms)

	// Init server
	var srv *Server
	if srv, err = NewServer(c, b, stg).Init(c); err != nil {
		astilog.Fatal(err)
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
