package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astitools/slice"
	"github.com/rs/xid"
	"github.com/rs/xlog"
)

// OS
const (
	OSLinux     = "linux"
	OSMaxOSX    = "macosx"
	OSWindows   = "windows"
	OSWindows32 = "windows_32"
	repoName    = "github.com/asticode/go-astichat"
)

// Builder represents a builder
type Builder struct {
	Logger               xlog.Logger
	pathWorkingDirectory string
	serverAddr           string
}

// New returns a new builder
func New(c Configuration) *Builder {
	return &Builder{
		Logger:               xlog.NopLogger,
		pathWorkingDirectory: c.PathWorkingDirectory,
		serverAddr:           c.ServerAddr,
	}
}

// RandomID allows testing functions using it
var RandomID = func() string {
	return xid.New().String()
}

// ExecCmd allows testing functions using it
var ExecCmd = func(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

// Build builds the client
func (b *Builder) Build(os, username string, prvClient *astichat.PrivateKey, pubServer *astichat.PublicKey) (o string, err error) {
	// Retrieve git version
	var v []byte
	if v, err = b.gitVersion(); err != nil {
		return
	}

	// Init output path
	o = fmt.Sprintf("%s/%s", b.pathWorkingDirectory, RandomID())

	// Marshal client's private key
	var prvClientBytes []byte
	if prvClientBytes, err = prvClient.MarshalText(); err != nil {
		return
	}

	// Marshal server's public key
	var pubServerBytes []byte
	if pubServerBytes, err = pubServer.MarshalText(); err != nil {
		return
	}

	// Init ldflags
	var ldflags = []string{
		"-X main.ClientPrivateKey=" + string(prvClientBytes),
		"-X main.Server=" + b.serverAddr,
		"-X main.ServerPublicKey=" + string(pubServerBytes),
		"-X main.Username=" + username,
		"-X main.Version=" + string(v),
	}

	// Init cmd
	var cmd = exec.Command("go", "build", "-o", o, "-ldflags", strings.Join(ldflags, " "), repoName+"/client")
	cmd.Env = b.buildEnv(os)

	// Exec
	var co []byte
	if co, err = ExecCmd(cmd); err != nil {
		err = fmt.Errorf("%s: %s", err, string(co))
		return
	}
	return
}

// buildEnv returns the build environment variables
// TODO Test cross platform
func (b *Builder) buildEnv(outputOS string) (o []string) {
	o = []string{"GOPATH=" + os.Getenv("GOPATH"), "PATH=" + os.Getenv("PATH")}
	switch outputOS {
	case OSMaxOSX:
		o = append(o, "GOOS=darwin", "GOARCH=amd64")
	case OSWindows:
		o = append(o, "GOOS=windows", "GOARCH=amd64")
	case OSWindows32:
		o = append(o, "GOOS=windows", "GOARCH=386")
	default:
		o = append(o, "GOOS=linux", "GOARCH=amd64")
	}
	return
}

// GitVersion retrieves the project's git version
func (b *Builder) gitVersion() (o []byte, err error) {
	var cmd = exec.Command("git", "--git-dir", fmt.Sprintf("%s/src/%s/.git", os.Getenv("GOPATH"), repoName), "rev-parse", "HEAD")
	b.Logger.Debugf("Running %s", strings.Join(cmd.Args, " "))
	if o, err = ExecCmd(cmd); err != nil {
		return
	}
	o = bytes.TrimSpace(o)
	return
}

// IsValidOS checks whether the OS is valid for the builder
func IsValidOS(os string) bool {
	return astislice.InStringSlice(os, []string{OSLinux, OSMaxOSX, OSWindows, OSWindows32})
}
