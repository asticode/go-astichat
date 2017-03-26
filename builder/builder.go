package builder

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
)

// Builder represents a builder
type Builder struct {
	keyBits              int
	Logger               xlog.Logger
	pathRootProject      string
	pathWorkingDirectory string
}

// New returns a new builder
func New(c Configuration) *Builder {
	return &Builder{
		keyBits:              c.KeyBits,
		Logger:               xlog.NopLogger,
		pathRootProject:      c.PathRootProject,
		pathWorkingDirectory: c.PathWorkingDirectory,
	}
}

// GeneratePrivateKey generates an rsa private key with an optional passphrase
func (b *Builder) GeneratePrivateKey(passphrase string) (pk *rsa.PrivateKey, k []byte, err error) {
	// Generate RSA key
	if pk, err = rsa.GenerateKey(rand.Reader, b.keyBits); err != nil {
		return
	}

	// Convert it to pem
	var block = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}

	// Encrypt the pem
	if len(passphrase) > 0 {
		if block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(passphrase), x509.PEMCipherAES256); err != nil {
			return
		}
	}

	// Encode to memory
	k = pem.EncodeToMemory(block)
	return
}

// Build builds the client
func (b *Builder) Build(os string, privateKey []byte) (o string, err error) {
	// Retrieve git version
	var v []byte
	if v, err = b.gitVersion(); err != nil {
		return
	}

	// Init output path
	o = fmt.Sprintf("%s/%s", b.pathWorkingDirectory, xid.New().String())

	// Init ldflags
	var ldflags = fmt.Sprintf("-X main.PrivateKey=%s -X main.Version=%s", base64.StdEncoding.EncodeToString(privateKey), v)

	// Init cmd
	var cmd = exec.Command("go", "build", "-o", o, "-ldflags", ldflags, "github.com/asticode/go-astichat/client")
	cmd.Env = b.buildEnv(os)

	// Exec
	b.Logger.Debugf("Running %s", strings.Join(append(cmd.Env, cmd.Args...), " "))
	var co []byte
	if co, err = cmd.CombinedOutput(); err != nil {
		err = fmt.Errorf("%s: %s", err, string(co))
		return
	}
	return
}

// buildEnv returns the build environment variables
func (b *Builder) buildEnv(outputOS string) (o []string) {
	o = []string{"GOPATH=" + os.Getenv("GOPATH"), "PATH=" + os.Getenv("PATH")}
	switch outputOS {
	case OSMaxOSX:
		o = append(o, "GOOS=darwin", "GOARCH=386")
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
	var cmd = exec.Command("git", "--git-dir", fmt.Sprintf("%s/.git", b.pathRootProject), "rev-parse", "HEAD")
	b.Logger.Debugf("Running %s", strings.Join(cmd.Args, " "))
	if o, err = cmd.CombinedOutput(); err != nil {
		return
	}
	o = bytes.TrimSpace(o)
	return
}

// ValidOS checks whether the OS is valid for the builder
func ValidOS(os string) bool {
	return astislice.InStringSlice(os, []string{OSLinux, OSMaxOSX, OSWindows, OSWindows32})
}
