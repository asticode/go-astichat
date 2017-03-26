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

	// Init mongo storage
	var stg = astichat.NewStorageMongo(l, ms)

	// Switch on subcommand
	switch s {
	default:
		// Server
		if err := Serve(c, l, b, stg); err != nil {
			l.Fatal(err)
		}
	}
}
