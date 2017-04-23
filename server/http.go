package main

import (
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
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/template"
	"github.com/julienschmidt/httprouter"
)

// ServerHTTP represents an HTTP server
type ServerHTTP struct {
	addr       string
	builder    *builder.Builder
	logger     astilog.Logger
	pathStatic string
	storage    astichat.Storage
	templates  *template.Template
}

// NewServerHTTP creates a new HTTP server
func NewServerHTTP(l astilog.Logger, addr, pathStatic string, b *builder.Builder, stg astichat.Storage) *ServerHTTP {
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
	r.GET("/", s.HandleHomepageGET)
	r.POST("/download", s.HandleDownloadPOST)
	r.GET("/now", s.HandleNowGET)
	r.POST("/token", s.HandleTokenPOST)

	// Static files
	r.ServeFiles("/static/*filepath", http.Dir(s.pathStatic))

	// Serve
	s.logger.Debugf("Listening and serving on http://%s", s.addr)
	if err := http.ListenAndServe(s.addr, r); err != nil {
		s.logger.Fatal(err)
	}
	return
}

// HandleHomepageGET returns the homepage handler
func (s *ServerHTTP) HandleHomepageGET(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Execute template
	if err := s.templates.ExecuteTemplate(rw, "/homepage.html", nil); err != nil {
		s.logger.Errorf("%s while executing homepage GET template", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// processErrors processes errors
func (s *ServerHTTP) processErrors(rw http.ResponseWriter, errRequest, errServer *error, redirectURL string) {
	// Server error
	var msg string
	if *errServer != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		msg = "Unknown error"
	}

	// Request error
	if *errRequest != nil {
		rw.WriteHeader(http.StatusBadRequest)
		msg = (*errRequest).Error()
	}

	// Print error
	if msg != "" {
		if redirectURL != "" {
			rw.Write([]byte("<script>window.location = \"" + redirectURL + "?error=" + msg + "\"</script>"))
		} else {
			json.NewEncoder(rw).Encode(astichat.Body{Error: &astichat.BodyError{Message: msg}})
		}
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
func (srv *ServerHTTP) HandleDownloadPOST(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Process HTTP errors
	var errServer error
	var errRequest error
	defer srv.processErrors(rw, &errRequest, &errServer, "/")

	// Username is empty
	var username = r.FormValue("username")
	if len(username) == 0 {
		srv.logger.Error("Empty username")
		errRequest = errors.New("Please enter a username")
		return
	}

	// Password is empty
	var password = r.FormValue("password")
	if len(password) == 0 {
		srv.logger.Error("Empty password")
		errRequest = errors.New("Please enter a password")
		return
	}

	// Check whether client wants to upgrade
	var isUpgrade = r.FormValue("is_upgrade") == "1"

	// Validate username
	var c astichat.Chatterer
	if isUpgrade {
		// Token is empty
		var token = r.FormValue("token")
		if len(token) == 0 {
			srv.logger.Error("Empty token")
			errRequest = errors.New("Please enter a token")
			return
		}

		// Fetch chatterer
		if c, errServer = srv.storage.ChattererFetchByUsername(username); errServer != nil && errServer != astichat.ErrNotFoundInStorage {
			srv.logger.Errorf("%s while fetching chatterer by username %s", errServer, username)
			return
		} else if errServer == astichat.ErrNotFoundInStorage {
			errServer = nil
			srv.logger.Errorf("Invalid username %s", username)
			errRequest = errors.New("Invalid username")
			return
		}

		// Decode the token
		var t astichat.Token
		if t, errServer = astichat.DecodeToken(token, c.ServerPrivateKey); errServer != nil {
			srv.logger.Errorf("%s while decoding token %s", errServer, token)
			return
		}

		// Validate token
		if errServer = t.Validate(c); errServer != nil {
			srv.logger.Errorf("%s while validating token %s", errServer, token)
			return

		}
	} else {
		// Username is unique
		if _, errServer = srv.storage.ChattererFetchByUsername(username); errServer != nil && errServer != astichat.ErrNotFoundInStorage {
			srv.logger.Errorf("%s while fetching chatterer by username %s", errServer, username)
			return
		} else if errServer == nil {
			srv.logger.Errorf("Username %s is already used", username)
			errRequest = errors.New("Username is already used")
			return
		}
		errServer = nil
	}

	// OS is valid
	var outputOS = r.FormValue("os")
	if !builder.IsValidOS(outputOS) {
		srv.logger.Errorf("Invalid os %s", outputOS)
		errRequest = errors.New("Invalid OS")
		return
	}

	// Generate client's private key
	var prvClient *astichat.PrivateKey
	if prvClient, errServer = AstichatNewPrivateKey(password); errServer != nil {
		srv.logger.Errorf("%s while generating private key", errServer)
		return
	}

	// Get client's public key
	var pubClient *astichat.PublicKey
	if pubClient, errServer = prvClient.PublicKey(); errServer != nil {
		srv.logger.Errorf("%s while getting public key from rsa private key", errServer)
		return
	}

	// Generate server's private key
	var prvServer *astichat.PrivateKey
	if prvServer, errServer = AstichatNewPrivateKey(""); errServer != nil {
		srv.logger.Errorf("%s while generating private key", errServer)
		return
	}

	// Get server's public key
	var pubServer *astichat.PublicKey
	if pubServer, errServer = prvServer.PublicKey(); errServer != nil {
		srv.logger.Errorf("%s while getting public key from rsa private key", errServer)
		return
	}

	// Build client
	var outputPath string
	if outputPath, errServer = BuilderBuild(srv.builder, outputOS, username, prvClient, pubServer); errServer != nil {
		srv.logger.Errorf("%s while building client for os %s", errServer, outputOS)
		return
	}
	defer OSRemove(outputPath)

	// Create/Update chatterer
	if isUpgrade {
		c.ClientPublicKey = pubClient
		c.ServerPrivateKey = prvServer
		c.Token = ""
		c.TokenAt = time.Time{}
		if errServer = srv.storage.ChattererUpdate(c); errServer != nil {
			srv.logger.Errorf("%s while updating chatterer with username %s", errServer, username)
			return
		}
	} else {
		if _, errServer = srv.storage.ChattererCreate(username, pubClient, prvServer); errServer != nil {
			srv.logger.Errorf("%s while creating chatterer with username %s", errServer, username)
			return
		}
	}

	// Read file
	var b []byte
	if b, errServer = IOUtilReadFile(outputPath); errServer != nil {
		srv.logger.Errorf("%s while reading file %s", errServer, outputPath)
		return
	}

	// Set headers
	rw.Header().Set("Pragma", "public")
	rw.Header().Set("Cache-Control", "private")
	rw.Header().Set("Content-Disposition", "attachment; filename=astichat.exe")
	rw.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	rw.Header().Set("Content-Transfer-Encoding", "binary")
	rw.Header().Set("Content-Length", strconv.Itoa(len(b)))
	rw.Write(b)
}

// handle allows handling simple requests
func (srv *ServerHTTP) handle(rw http.ResponseWriter, r *http.Request, expectedMsg []byte, fn func(c astichat.Chatterer) ([]byte, error)) {
	// Process HTTP errors
	var errServer error
	var errRequest error
	defer srv.processErrors(rw, &errRequest, &errServer, "")

	// Unmarshal
	var b astichat.Body
	if errServer = json.NewDecoder(r.Body).Decode(&b); errServer != nil {
		srv.logger.Errorf("%s while unmarshaling body", errServer)
		return
	}

	// Retrieve chatterer
	var c astichat.Chatterer
	if c, errServer = srv.storage.ChattererFetchByUsername(b.Request.Username); errServer != nil {
		srv.logger.Errorf("%s while fetching chatterer by username %s", errServer, b.Request.Username)
		return
	}

	// Process body
	var msg []byte
	if msg, errServer = b.Process(astichat.TimeNow(), c.ServerPrivateKey); errServer != nil {
		srv.logger.Errorf("%s while processing body", errServer)
		return
	}

	// Validate message
	if errServer = astichat.ValidateMessage(msg, expectedMsg); errServer != nil {
		srv.logger.Errorf("%s while validating message", errServer)
		return
	}

	// Custom handler
	if msg, errServer = fn(c); errServer != nil {
		srv.logger.Errorf("%s while executing custom handler", errServer)
		return
	}

	// Create new body
	if b, errServer = astichat.NewBody(msg, astichat.TimeNow(), "", c.ClientPublicKey); errServer != nil {
		srv.logger.Errorf("%s while creating new body", errServer)
		return
	}

	// Write
	if errServer = json.NewEncoder(rw).Encode(b); errServer != nil {
		srv.logger.Errorf("%s while writing", errServer)
		return
	}
}

// HandleNowGET returns the current time
// It shouldn't be protected as we need it to protect other messages
func (srv *ServerHTTP) HandleNowGET(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Process HTTP errors
	var errServer error
	var errRequest error
	defer srv.processErrors(rw, &errRequest, &errServer, "")

	// Marshal
	if errServer = json.NewEncoder(rw).Encode(astichat.TimeNow()); errServer != nil {
		srv.logger.Errorf("%s while writing", errServer)
		return
	}
}

// HandleTokenPOST delivers a token for a specific username that can be used during a short period of time to interact
// with the server
func (srv *ServerHTTP) HandleTokenPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	srv.handle(rw, r, astichat.MessageToken, func(c astichat.Chatterer) (b []byte, err error) {
		// Generate token
		c.Token = astichat.GenerateToken()
		c.TokenAt = astichat.TimeNow()

		// Store token
		if err = srv.storage.ChattererUpdate(c); err != nil {
			srv.logger.Errorf("%s while updating chatterer %s", err, c.ID)
			return
		}
		b = []byte(c.Token)
		return
	})
}
