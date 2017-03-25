package main

import (
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
)

// Flags
var (
	outputOS   = flag.String("os", "linux", "the OS for which to build")
	outputPath = flag.String("o", ".", "the output path")
	passphrase = flag.String("p", "", "your passphrase")
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
	var b = NewBuilder(c, l)
	defer b.Close()

	// Switch on subcommand
	var err error
	switch s {
	default:
		// Empty passphrase
		if len(*passphrase) == 0 {
			l.Fatal("Missing passphrase (use the -p option)")
		}

		// Generate key
		var k []byte
		if k, err = b.GenerateKey(*passphrase); err != nil {
			l.Fatalf("%s while generating key")
		}

		// Build
		if err = b.Build(*outputPath, *outputOS, k); err != nil {
			l.Fatalf("%s while building", err)
		}

	}
}
