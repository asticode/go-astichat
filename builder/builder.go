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

	"github.com/rs/xlog"
)

// OS
const (
	OSLinux     = "linux"
	OSMaxOSX    = "macosx"
	OSWindows   = "windows"
	OSWindows64 = "windows_64"
)

// Builder represents a builder
type Builder struct {
	keyBits         int
	logger          xlog.Logger
	rootProjectPath string
}

// New returns a new builder
func New(keyBits int, l xlog.Logger, rootProjectPath string) *Builder {
	l.Debug("Starting builder")
	return &Builder{
		keyBits:         keyBits,
		logger:          l,
		rootProjectPath: rootProjectPath,
	}
}

// Close closes the builder
func (b *Builder) Close() {
	b.logger.Debug("Stopping client")
}

// GenerateKey generates an rsa key with an optional passphrase
func (b *Builder) GenerateKey(passphrase string) (k []byte, err error) {
	// Generate RSA key
	var pk *rsa.PrivateKey
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
func (b *Builder) Build(outputPath, outputOS string, privateKey []byte) (err error) {
	// Retrieve git version
	var v []byte
	if v, err = b.GitVersion(); err != nil {
		return
	}

	// Build
	var ldflags = fmt.Sprintf("-X main.PrivateKey=%s -X main.Version=%s", base64.StdEncoding.EncodeToString(privateKey), v)
	var cmd = exec.Command("go", "build", "-o", fmt.Sprintf("%s/client/client", outputPath), "-ldflags", ldflags, fmt.Sprintf("%s/client", b.rootProjectPath))
	cmd.Env = b.buildEnv(outputOS)
	b.logger.Debugf("Running %s", strings.Join(append(cmd.Env, cmd.Args...), " "))
	var o []byte
	if o, err = cmd.CombinedOutput(); err != nil {
		err = fmt.Errorf("%s: %s", err, string(o))
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
		o = append(o, "GOOS=windows", "GOARCH=386")
	case OSWindows64:
		o = append(o, "GOOS=windows", "GOARCH=amd64")
	default:
		o = append(o, "GOOS=linux", "GOARCH=amd64")
	}
	return
}

// GitVersion retrieves the project's git version
func (b *Builder) GitVersion() (o []byte, err error) {
	var cmd = exec.Command("git", "--git-dir", fmt.Sprintf("%s/.git", b.rootProjectPath), "rev-parse", "HEAD")
	b.logger.Debugf("Running %s", strings.Join(cmd.Args, " "))
	if o, err = cmd.CombinedOutput(); err != nil {
		return
	}
	o = bytes.TrimSpace(o)
	return
}
