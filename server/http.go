package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	"github.com/asticode/go-astitools/template"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/xid"
	"github.com/rs/xlog"
)

// Constants
const (
	contextKeyBuilder   = "builder"
	contextKeyLogger    = "logger"
	contextKeyStorage   = "storage"
	contextKeyTemplates = "templates"
)

// ServerHTTP represents an HTTP server
type ServerHTTP struct {
	addr       string
	builder    *builder.Builder
	logger     xlog.Logger
	pathStatic string
	storage    astichat.Storage
	templates  *template.Template
}

// NewServerHTTP creates a new HTTP server
func NewServerHTTP(l xlog.Logger, addr, pathStatic string, b *builder.Builder, stg astichat.Storage) *ServerHTTP {
	return &ServerHTTP{
		addr:       addr,
		builder:    b,
		logger:     l,
		pathStatic: pathStatic,
		storage:    stg,
	}
}

// Init initializes the HTTP server
func (s *ServerHTTP) Init(c Configuration) (err error) {
	// Parse templates
	if s.templates, err = astitemplate.ParseDirectory(c.PathTemplates, ".html"); err != nil {
		return
	}
	return
}

// ListenAndServe listens and serve
func (s *ServerHTTP) ListenAndServe() {
	// Init router
	var r = httprouter.New()

	// Website
	r.GET("/", HandleHomepageGET)
	r.POST("/download", HandleDownloadPOST)
	r.GET("/now", HandleNowGET)
	r.POST("/token", HandleTokenPOST)

	// Static files
	r.ServeFiles("/static/*filepath", http.Dir(s.pathStatic))

	// Serve
	s.logger.Debugf("Listening and serving on http://%s", s.addr)
	if err := http.ListenAndServe(s.addr, s.AdaptHandler(r)); err != nil {
		s.logger.Fatal(err)
	}
	return
}

// AdaptHandle adapts a handler
func (s *ServerHTTP) AdaptHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		r = r.WithContext(NewContextWithBuilder(r.Context(), s.builder))
		r = r.WithContext(NewContextWithLogger(r.Context(), s.logger))
		r = r.WithContext(NewContextWithStorage(r.Context(), s.storage))
		r = r.WithContext(NewContextWithTemplates(r.Context(), s.templates))
		h.ServeHTTP(rw, r)
	})
}

// NewContextWithBuilder creates a context with the builder
func NewContextWithBuilder(ctx context.Context, b *builder.Builder) context.Context {
	// Parse templates
	return context.WithValue(ctx, contextKeyBuilder, b)
}

// BuilderFromContext retrieves the builder from the context
func BuilderFromContext(ctx context.Context) *builder.Builder {
	if t, ok := ctx.Value(contextKeyBuilder).(*builder.Builder); ok {
		return t
	}
	return &builder.Builder{}
}

// NewContextWithLogger creates a context with the logger
func NewContextWithLogger(ctx context.Context, l xlog.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, l)
}

// LoggerFromContext retrieves the logger from the context
func LoggerFromContext(ctx context.Context) xlog.Logger {
	if l, ok := ctx.Value(contextKeyLogger).(xlog.Logger); ok {
		return l
	}
	return xlog.NopLogger
}

// NewContextWithStorage creates a context with the storage
func NewContextWithStorage(ctx context.Context, s astichat.Storage) context.Context {
	// Parse templates
	return context.WithValue(ctx, contextKeyStorage, s)
}

// StorageFromContext retrieves the storage from the context
func StorageFromContext(ctx context.Context) astichat.Storage {
	if t, ok := ctx.Value(contextKeyStorage).(astichat.Storage); ok {
		return t
	}
	return astichat.NopStorage{}
}

// NewContextWithTemplates creates a context with the templates
func NewContextWithTemplates(ctx context.Context, t *template.Template) context.Context {
	return context.WithValue(ctx, contextKeyTemplates, t)
}

// TemplatesFromContext retrieves the templates from the context
func TemplatesFromContext(ctx context.Context) *template.Template {
	if t, ok := ctx.Value(contextKeyTemplates).(*template.Template); ok {
		return t
	}
	return &template.Template{}
}

// HandleHomepageGET returns the homepage handler
func HandleHomepageGET(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Init
	var l = LoggerFromContext(r.Context())
	var t = TemplatesFromContext(r.Context())

	// Execute template
	if err := t.ExecuteTemplate(rw, "/homepage.html", nil); err != nil {
		l.Errorf("%s while executing homepage GET template", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// ProcessHTTPError processes HTTP errors
func ProcessHTTPErrors(rw http.ResponseWriter, l xlog.Logger, errRequest, errServer *error) {
	// Request error
	if *errRequest != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("<script>window.location = '/?error=" + (*errRequest).Error() + "'</script>"))
		return
	}

	// Server error
	if *errServer != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// AstichatNewPrivateKey allows testing functions using it
var AstichatNewPrivateKey = func(passphrase string) (*astichat.PrivateKey, error) {
	return astichat.NewPrivateKey(passphrase)
}

// BuilderBuild allows testing functions using it
var BuilderBuild = func(b *builder.Builder, os, username string, prvClient *astichat.PrivateKey, pubServer *astichat.PublicKey) (string, error) {
	return b.Build(os, username, prvClient, pubServer)
}

// OSRemove allows testing functions using it
var OSRemove = func(path string) error {
	return os.Remove(path)
}

// IOUtilReadFile allows testing functions using it
var IOUtilReadFile = func(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// HandleDownloadPOST returns a newly built client
// TODO Present file as inline attachment even in AJAX => split in 2 steps?
// TODO Regenerate private key on upgrade. To make sure upgrade demand comes from the right place, client must
// ask for a token generated server-side (and stored in the storage with a timestamp), and on upgrade the server
// checks the encrypted message contains the correct token and validate the timestamp as well
func HandleDownloadPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Init
	var bd = BuilderFromContext(r.Context())
	var l = LoggerFromContext(r.Context())
	var s = StorageFromContext(r.Context())

	// Process HTTP errors
	var errServer error
	var errRequest error
	defer ProcessHTTPErrors(rw, l, &errRequest, &errServer)

	// Username is empty
	var username = r.FormValue("username")
	if len(username) == 0 {
		l.Error("Empty username")
		errRequest = errors.New("Please enter a username")
		return
	}

	// Username is unique
	if _, errServer = s.ChattererFetchByUsername(username); errServer != nil && errServer != astichat.ErrNotFoundInStorage {
		l.Errorf("%s while fetching chatterer by username %s", errServer, username)
		return
	} else if errServer == nil {
		l.Errorf("Username %s is already used", username)
		errRequest = errors.New("Username is already used")
		return
	}
	errServer = nil

	// Password is not empty
	var password = r.FormValue("password")
	if len(password) == 0 {
		l.Error("Empty password")
		errRequest = errors.New("Please enter a password")
		return
	}

	// OS is valid
	var outputOS = r.FormValue("os")
	if !builder.IsValidOS(outputOS) {
		l.Errorf("Invalid os %s", outputOS)
		errRequest = errors.New("Invalid OS")
		return
	}

	// Generate client's private key
	var prvClient *astichat.PrivateKey
	if prvClient, errServer = AstichatNewPrivateKey(password); errServer != nil {
		l.Errorf("%s while generating private key", errServer)
		return
	}

	// Get client's public key
	var pubClient *astichat.PublicKey
	if pubClient, errServer = prvClient.PublicKey(); errServer != nil {
		l.Errorf("%s while getting public key from rsa private key", errServer)
		return
	}

	// Generate server's private key
	// TODO Add passphrase too ? For unmarshal, use global variable taken from conf ?
	var prvServer *astichat.PrivateKey
	if prvServer, errServer = AstichatNewPrivateKey(""); errServer != nil {
		l.Errorf("%s while generating private key", errServer)
		return
	}

	// Get server's public key
	var pubServer *astichat.PublicKey
	if pubServer, errServer = prvServer.PublicKey(); errServer != nil {
		l.Errorf("%s while getting public key from rsa private key", errServer)
		return
	}

	// Build client
	var outputPath string
	if outputPath, errServer = BuilderBuild(bd, outputOS, username, prvClient, pubServer); errServer != nil {
		l.Errorf("%s while building client for os %s", errServer, outputOS)
		return
	}
	defer OSRemove(outputPath)

	// Create chatterer
	if _, errServer = s.ChattererCreate(username, pubClient, prvServer); errServer != nil {
		l.Errorf("%s while creating chatterer with username %s and public key %s", errServer, username, pubClient)
		return
	}

	// Read file
	var b []byte
	if b, errServer = IOUtilReadFile(outputPath); errServer != nil {
		l.Errorf("%s while reading file %s", errServer, outputPath)
		return
	}

	// Set headers
	rw.Header().Set("Pragma", "public")
	rw.Header().Set("Cache-Control", "private")
	rw.Header().Set("Content-Disposition", "attachment; filename=astichat.exe")
	rw.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	rw.Header().Set("Content-Transfer-Encodin", "binary")
	rw.Header().Set("Content-Length", strconv.Itoa(len(b)))
	rw.Write(b)
}

// HandleNowGET returns the current time
func HandleNowGET(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Init
	var l = LoggerFromContext(r.Context())

	// Marshal
	var err error
	if err = json.NewEncoder(rw).Encode(Now()); err != nil {
		l.Errorf("%s while json marshaling now", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GenerateToken allows testing functions using it
var GenerateToken = func() string {
	return xid.New().String()
}

// Now allows testing functions using it
var Now = func() time.Time {
	return time.Now()
}

// HandleTokenPOST delivers a token for a specific username that can be used during a short period of time to interact
// with the server. If the username doesn't exist, a token is still generated to avoid allowing recreating the list of
// usernames with a simple script.
func HandleTokenPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Init
	var l = LoggerFromContext(r.Context())
	var s = StorageFromContext(r.Context())

	// Process HTTP errors
	var errServer error
	var errRequest error
	defer ProcessHTTPErrors(rw, l, &errRequest, &errServer)

	// Username is empty
	var username = r.FormValue("username")
	if len(username) == 0 {
		l.Error("Empty username")
		errRequest = errors.New("Please enter a username")
		return
	}

	// Fetch chatterer
	var c astichat.Chatterer
	if c, errServer = s.ChattererFetchByUsername(username); errServer != nil && errServer != astichat.ErrNotFoundInStorage {
		l.Errorf("%s while fetching chatterer by username %s", errServer, username)
		return
	}

	// Generate token
	c.Token = GenerateToken()
	c.TokenAt = Now()

	// Return token even though username doesn't exist
	if errServer != nil {
		l.Errorf("Username %s doesn't exist", username)
		rw.Write([]byte(c.Token))
		errServer = nil
		return
	}

	// Store token
	if errServer = s.ChattererUpdate(c); errServer != nil {
		l.Errorf("%s while updating chatterer %s", errServer, c.ID.Hex())
		return
	}
}
