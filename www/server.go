package main

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/asticode/go-astichat/astichat"
	"github.com/asticode/go-astichat/builder"
	"github.com/asticode/go-astitools/template"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
)

// Constants
const (
	contextKeyBuilder   = "builder"
	contextKeyLogger    = "logger"
	contextKeyStorage   = "storage"
	contextKeyTemplates = "templates"
)

// Serve initialize an HTTP server
func Serve(c Configuration, l xlog.Logger, b *builder.Builder, s astichat.Storage) (err error) {
	// Init router
	var r = httprouter.New()

	// Website
	r.GET("/", HandleHomepageGET)
	r.POST("/download/client", HandleDownloadClientGET)

	// Static files
	r.ServeFiles("/static/*filepath", http.Dir(c.PathStatic))

	// Parse templates
	var t *template.Template
	if t, err = astitemplate.ParseDirectory(c.PathTemplates, ".html"); err != nil {
		return
	}

	// Serve
	l.Debugf("Listening and serving on %s", c.ListenAddr)
	if err = http.ListenAndServe(c.ListenAddr, AdaptHandler(r, l, b, s, t)); err != nil {
		return
	}
	return
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

// AdaptHandle adapts a handler
func AdaptHandler(h http.Handler, l xlog.Logger, b *builder.Builder, s astichat.Storage, t *template.Template) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		r = r.WithContext(NewContextWithBuilder(r.Context(), b))
		r = r.WithContext(NewContextWithLogger(r.Context(), l))
		r = r.WithContext(NewContextWithStorage(r.Context(), s))
		r = r.WithContext(NewContextWithTemplates(r.Context(), t))
		h.ServeHTTP(rw, r)
	})
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

// HandleDownloadClientGET returns the download client handler
func HandleDownloadClientGET(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Init
	var bd = BuilderFromContext(r.Context())
	var l = LoggerFromContext(r.Context())
	var s = StorageFromContext(r.Context())

	// Username is not empty
	var username = r.FormValue("username")
	if len(username) == 0 {
		l.Error("Empty username")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Please enter a username"))
		return
	}

	// Username is unique
	var err error
	if _, err = s.ChattererFetchByUsername(username); err != nil && err != astichat.ErrNotFoundInDB {
		l.Errorf("%s while fetching peer by username %s", err, username)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	} else if err == nil {
		l.Errorf("Username %s is already used", username)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(append([]byte(fmt.Sprintf("Username %s is already used", username))))
		return
	}

	// Password is not empty
	var password = r.FormValue("password")
	if len(password) == 0 {
		l.Error("Empty password")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Please enter a password"))
		return
	}

	// OS is valid
	var outputOS = r.FormValue("os")
	if !builder.ValidOS(outputOS) {
		l.Errorf("Invalid os %s", outputOS)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf("Invalid os %s", outputOS)))
		return
	}

	// TODO Handle given public keys

	// Generate key
	var pk *rsa.PrivateKey
	var b []byte
	if pk, b, err = bd.GeneratePrivateKey(password); err != nil {
		l.Errorf("%s while generating private key with password %s", err, password)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Init public key
	var pub astichat.PublicKey
	if pub, err = astichat.NewPublicKeyFromRSAPrivateKey(pk); err != nil {
		l.Errorf("%s while creating public key from rsa private key", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create chatterer
	if _, err = s.ChattererCreate(username, pub); err != nil {
		l.Errorf("%s while creating chatterer with username %s and public key %s", err, username, pub)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Build client
	var outputPath string
	if outputPath, err = bd.Build(outputOS, b); err != nil {
		l.Errorf("%s while building client for os %s", err, outputOS)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputPath)

	// Read file
	if b, err = ioutil.ReadFile(outputPath); err != nil {
		l.Errorf("%s while reading file %s", err, outputPath)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers
	rw.Header().Set("Content-Disposition", "attachment; filename=astichat.exe")
	rw.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	rw.Header().Set("Content-Length", strconv.Itoa(len(b)))
	rw.Write(b)
}
