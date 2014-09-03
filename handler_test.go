package gist_test

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benbjohnson/gist"
	"github.com/gorilla/sessions"
)

// Ensure the user gets redirected to be authorized.
func TestHandler_Root_Unauthorized(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	var redirectURI string
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectURI = req.URL.RequestURI()
			return errors.New("no redirects")
		},
	}

	resp, _ := client.Get(h.Server.URL)
	resp.Body.Close()
	equals(t, "/_/authorize", redirectURI)
}

// Ensure the user sees a list of their gists when authorized.
func TestHandler_Root_Authorized(t *testing.T) {
	// Create an authenticated user.
	store := NewTestStore()
	store.GetFunc = func(r *http.Request, name string) (*sessions.Session, error) {
		return &sessions.Session{Values: map[interface{}]interface{}{"UserID": 1000}}, nil
	}

	// Return a single gist.
	client := &MockGitHubClient{
		GistsFunc: func(username string) ([]*gist.Gist, error) {
			return []*gist.Gist{
				&gist.Gist{ID: "abc", Description: "my gist"},
			}, nil
		},
	}

	// Setup handler.
	h := NewTestHandler()
	h.Handler.Store = store
	h.Handler.NewGitHubClient = func(_ string) gist.GitHubClient { return client }
	defer h.Close()

	// Retrieve root.
	resp, err := http.Get(h.Server.URL)
	ok(t, err)
	assert(t, strings.Contains(readall(resp.Body), "my gist"), "expected substring")
}

// Ensure the user is redirected to GitHub for authorization.
func TestHandler_Authorize(t *testing.T) { t.Skip("pending") }

// Ensure the OAuth2 callback is processed correctly.
func TestHandler_Authorized(t *testing.T) { t.Skip("pending") }

// Ensure an OAuth2 callback with non-matching auth state causes an error.
func TestHandler_Authorized_ErrInvalidState(t *testing.T) { t.Skip("pending") }

// Ensure an oEmbed is processed correctly.
func TestHandler_OEmbed(t *testing.T) { t.Skip("pending") }

// Ensure an oEmbed with width/height set is returned correctly.
func TestHandler_OEmbed_WidthHeight(t *testing.T) { t.Skip("pending") }

// Ensure a missing gist returns an error.
func TestHandler_OEmbed_ErrNotFound(t *testing.T) { t.Skip("pending") }

// Ensure an XML oEmbed returns an error.
func TestHandler_OEmbed_XML_ErrStatusNotImplemented(t *testing.T) { t.Skip("pending") }

// Ensure a gist can be retrieved.
func TestHandler_Gist(t *testing.T) { t.Skip("pending") }

// Ensure a gist with a non-canonical URL is redirected.
func TestHandler_Gist_NonCanonical(t *testing.T) { t.Skip("pending") }

// Ensure a gist will be reloaded if user is authorized.
func TestHandler_Gist_Reload(t *testing.T) { t.Skip("pending") }

// Ensure a gist with an invalid path returns an error.
func TestHandler_Gist_ErrInvalidPath(t *testing.T) { t.Skip("pending") }

// Ensure a path is correctly parsed into gist id and filename.
func TestHandler_ParsePath(t *testing.T) { t.Skip("pending (TT)") }

// TestHandler represents a handler used for testing.
type TestHandler struct {
	*gist.Handler
	Path   string
	DB     *gist.DB
	Server *httptest.Server
}

// NewTestHandler returns a new instance of TestHandler.
func NewTestHandler() *TestHandler {
	path := tempfile()
	os.Mkdir(path, 0700)

	// Open database.
	db := &gist.DB{}
	db.GistPath = filepath.Join(path, "gists")
	if err := db.Open(filepath.Join(path, "db"), 0600); err != nil {
		log.Fatal(err)
	}

	// Create a test user.
	err := db.Update(func(tx *gist.Tx) error {
		return tx.SaveUser(&gist.User{ID: 1000, Username: "benbjohnson", AccessToken: "XYZ"})
	})
	if err != nil {
		log.Fatal(err)
	}

	// Open handler and test HTTP server.
	h := gist.NewHandler(db, "ABC", "123")
	s := httptest.NewServer(h)

	return &TestHandler{h, path, db, s}
}

func (h *TestHandler) Close() {
	h.Server.Close()
	h.DB.Close()
	os.RemoveAll(h.Path)
}

// TestStore represents a mockable session store.
type TestStore struct {
	GetFunc  func(r *http.Request, name string) (*sessions.Session, error)
	NewFunc  func(r *http.Request, name string) (*sessions.Session, error)
	SaveFunc func(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

// NewTestStore returns a new instance of NewTestStore.
func NewTestStore() *TestStore {
	return &TestStore{}
}

func (s *TestStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return s.GetFunc(r, name)
}

func (s *TestStore) New(r *http.Request, name string) (*sessions.Session, error) {
	return s.NewFunc(r, name)
}

func (s *TestStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return s.SaveFunc(r, w, session)
}

func readall(r io.Reader) string {
	b, _ := ioutil.ReadAll(r)
	return string(b)
}
