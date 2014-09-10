package gist_test

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.google.com/p/goauth2/oauth"
	"github.com/benbjohnson/gist"
	"github.com/gorilla/sessions"
)

// Ensure the user sees the home page.
func TestHandler_Root_Unauthorized(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	resp, _ := http.Get(h.Server.URL)
	resp.Body.Close()
	equals(t, 200, resp.StatusCode)
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
func TestHandler_Authorize(t *testing.T) {
	// Create the mock session store.
	var saved bool
	store := NewTestStore()
	session := sessions.NewSession(store, "")
	store.GetFunc = func(r *http.Request, name string) (*sessions.Session, error) {
		return session, nil
	}
	store.SaveFunc = func(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
		saved = true
		return nil
	}

	// Setup handler.
	h := NewTestHandler()
	h.Handler.Store = store
	defer h.Close()

	// Create non-redirecting client.
	var redirectURL *url.URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectURL = req.URL
			return errors.New("no redirects")
		},
	}

	// Retrieve authorize redirect.
	// We should be redirected to GitHub's OAuth URL.
	// We should save the auth state to the session so it can be check on callback.
	resp, _ := client.Get(h.Server.URL + "/_/login")
	resp.Body.Close()
	equals(t, "https", redirectURL.Scheme)
	equals(t, "github.com", redirectURL.Host)
	equals(t, "/login/oauth/authorize", redirectURL.Path)
	equals(t, 32, len(redirectURL.Query().Get("state")))

	assert(t, saved, "expected session save")
	equals(t, redirectURL.Query().Get("state"), session.Values["AuthState"])
}

// Ensure the OAuth2 callback is processed correctly.
func TestHandler_Authorized(t *testing.T) {
	// Create the mock session store.
	store := NewTestStore()
	session := sessions.NewSession(store, "")
	session.Values["AuthState"] = "abc123"
	store.GetFunc = func(r *http.Request, name string) (*sessions.Session, error) {
		return session, nil
	}
	store.SaveFunc = func(r *http.Request, w http.ResponseWriter, session *sessions.Session) error { return nil }

	// Return a fake user.
	client := &MockGitHubClient{}
	client.GistsFunc = func(username string) ([]*gist.Gist, error) { return nil, nil }
	client.UserFunc = func(username string) (*gist.User, error) {
		return &gist.User{ID: 1000, Username: "john"}, nil
	}

	// Create non-redirecting client.
	var redirectURL *url.URL
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectURL = req.URL
			return errors.New("no redirects")
		},
	}

	// Setup handler.
	h := NewTestHandler()
	h.Handler.Store = store
	h.Handler.NewGitHubClient = func(token string) gist.GitHubClient {
		equals(t, "mytoken", token)
		return client
	}
	h.ExchangeFunc = func(code string) (*oauth.Token, error) { return &oauth.Token{AccessToken: "mytoken"}, nil }
	defer h.Close()

	// Process callback.
	resp, _ := httpClient.Get(h.Server.URL + "/_/login/callback?state=abc123")
	resp.Body.Close()

	// We should be redirected to the root path.
	equals(t, 302, resp.StatusCode)
	equals(t, "/_/dashboard", redirectURL.Path)

	// The session should have the user id set.
	equals(t, 1000, session.Values["UserID"])

	// The user should exist.
	h.DB.View(func(tx *gist.Tx) error {
		u, _ := tx.User(1000)
		equals(t, &gist.User{ID: 1000, Username: "john", AccessToken: "mytoken"}, u)
		return nil
	})
}

// Ensure an oEmbed is processed correctly.
func TestHandler_OEmbed(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	// Create the gist in the database.
	h.DB.Update(func(tx *gist.Tx) error {
		return tx.SaveGist(&gist.Gist{ID: "abc123", UserID: 1000, Description: "My Gist"})
	})

	// Retrieve oEmbed.
	u, _ := url.Parse(h.Server.URL + "/oembed.json")
	u.RawQuery = (&url.Values{"url": {"https://gist.exposed/benbjohnson/abc123"}}).Encode()
	resp, err := http.Get(u.String())
	ok(t, err)
	equals(t, 200, resp.StatusCode)

	html, _ := json.Marshal(`<div class="gist-exposed" style="position: relative; padding-bottom: 300; padding-top: 0px; height: 0; overflow: hidden;"><iframe style="position: absolute; top:0; left: 0; width: 100%; height: 100%; border: none;" src="https://gist.exposed/benbjohnson/abc123/"></iframe></div>`)
	equals(t, `{"version":"1.0","type":"rich","html":`+string(html)+`,"height":300,"title":"My Gist","provider_name":"Gist Exposed!","provider_url":"https://gist.exposed"}`+"\n", readall(resp.Body))
}

// Ensure an oEmbed with width/height set is returned correctly.
func TestHandler_OEmbed_WidthHeight(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	// Create the gist in the database.
	h.DB.Update(func(tx *gist.Tx) error {
		return tx.SaveGist(&gist.Gist{ID: "abc123", UserID: 1000, Description: "My Gist"})
	})

	// Retrieve oEmbed.
	u, _ := url.Parse(h.Server.URL + "/oembed.json")
	u.RawQuery = (&url.Values{"url": {"https://gist.exposed/benbjohnson/abc123?width=50&height=60"}}).Encode()
	resp, err := http.Get(u.String())
	ok(t, err)
	equals(t, 200, resp.StatusCode)

	// Unmarshal and check width & height.
	var o struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	b, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &o)
	equals(t, 50, o.Width)
	equals(t, 60, o.Height)
}

// Ensure an oEmbed for a missing gist returns a 404.
func TestHandler_OEmbed_ErrNotFound(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	// Retrieve oEmbed.
	u, _ := url.Parse(h.Server.URL + "/oembed.json")
	u.RawQuery = (&url.Values{"url": {"https://gist.exposed/benbjohnson/abc123"}}).Encode()
	resp, err := http.Get(u.String())
	resp.Body.Close()
	ok(t, err)
	equals(t, 404, resp.StatusCode)
}

// Ensure an oEmbed with a bad url returns a 404.
func TestHandler_OEmbed_ErrInvalidPath(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	// Retrieve oEmbed.
	resp, err := http.Get(h.Server.URL + "/oembed.json?url=bad_url")
	resp.Body.Close()
	ok(t, err)
	equals(t, 404, resp.StatusCode)
}

// Ensure an XML oEmbed returns an error.
func TestHandler_OEmbed_XML_ErrStatusNotImplemented(t *testing.T) {
	h := NewTestHandler()
	defer h.Close()

	// Retrieve oEmbed.
	resp, _ := http.Get(h.Server.URL + "/oembed.xml")
	resp.Body.Close()
	equals(t, 501, resp.StatusCode)
}

// Ensure a gist can be retrieved.
func TestHandler_Gist_Authorized(t *testing.T) {
	// Run mock GitHub raw server.
	mux := http.NewServeMux()
	mux.HandleFunc(`/index.html`, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body></body></html>`))
	})
	mux.HandleFunc(`/awesome.js`, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`alert(100);`))
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	// Create the mock session store.
	store := NewTestStore()
	session := sessions.NewSession(store, "")
	session.Values["UserID"] = 1000
	store.GetFunc = func(r *http.Request, name string) (*sessions.Session, error) { return session, nil }
	store.SaveFunc = func(r *http.Request, w http.ResponseWriter, session *sessions.Session) error { return nil }

	// Return a gist data.
	client := &MockGitHubClient{}
	client.GistFunc = func(id string) (*gist.Gist, error) {
		equals(t, "xxx", id)
		return &gist.Gist{
			ID:          "xxx",
			UserID:      1000,
			Description: "my gist",
			Public:      true,
			URL:         "-",
			Files: []*gist.GistFile{
				&gist.GistFile{Size: 100, Filename: "index.html", RawURL: s.URL + "/index.html"},
				&gist.GistFile{Size: 200, Filename: "awesome.js", RawURL: s.URL + "/awesome.js"},
			},
		}, nil
	}

	// Setup handler.
	h := NewTestHandler()
	h.Handler.Store = store
	h.Handler.NewGitHubClient = func(token string) gist.GitHubClient { return client }
	h.DB.NewGitHubClient = h.Handler.NewGitHubClient
	defer h.Close()

	// Process callback.
	resp, _ := http.Get(h.Server.URL + "/john/xxx/")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// The HTML file should be returned.
	equals(t, 200, resp.StatusCode)
	equals(t, `<html><body></body></html>`, string(body))

	// The files should be saved to the hard drive.
	var content []byte
	content, _ = ioutil.ReadFile(filepath.Join(h.DB.GistPath, "xxx", "index.html"))
	equals(t, `<html><body></body></html>`, string(content))
	content, _ = ioutil.ReadFile(filepath.Join(h.DB.GistPath, "xxx", "awesome.js"))
	equals(t, `alert(100);`, string(content))

	// The gist should be saved to the db.
	h.DB.View(func(tx *gist.Tx) error {
		g, _ := tx.Gist("xxx")
		assert(t, g != nil, "expected gist")
		return nil
	})
}

// Ensure a path is correctly parsed into gist id and filename.
func TestParsePath(t *testing.T) {
	var tests = []struct {
		path     string
		gistID   string
		filename string
		err      string
	}{
		{path: "/", gistID: "", filename: "", err: "invalid path"},
		{path: "/abc123", gistID: "abc123", filename: "", err: "non-canonical path"},
		{path: "/abc123/", gistID: "abc123", filename: "", err: ""},
		{path: "/abc123/index.html", gistID: "abc123", filename: "index.html", err: ""},
		{path: "/user100/abc123", gistID: "abc123", filename: "", err: "non-canonical path"},
		{path: "/user100/abc123/", gistID: "abc123", filename: "", err: ""},
		{path: "/user100/abc123/index.html", gistID: "abc123", filename: "index.html", err: ""},
		{path: "/user100/abc123/subdir/index.html", gistID: "", filename: "", err: "invalid path: /user100/abc123/subdir/index.html"},
	}
	for i, tt := range tests {
		gistID, filename, err := gist.ParsePath(tt.path)
		var errstr string
		if err != nil {
			errstr = err.Error()
		}
		if tt.err != errstr {
			t.Errorf("%d. error: exp: %s, got: %s", i, tt.err, errstr)
		} else if tt.gistID != gistID {
			t.Errorf("%d. gistID: exp: %s, got: %s", i, tt.gistID, gistID)
		} else if tt.filename != filename {
			t.Errorf("%d. filename: exp: %s, got: %s", i, tt.filename, filename)
		}
	}
}

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
	if !testing.Verbose() {
		h.Logger = log.New(ioutil.Discard, "", 0)
	}
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
