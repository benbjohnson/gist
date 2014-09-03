package gist_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benbjohnson/gist"
)

// Ensure that the GitHub client can retrieve a user by username.
func TestGitHub_User(t *testing.T) {
	// Create mock GitHub API server.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/users/john", r.URL.Path)
		fmt.Fprint(w, `{"login": "john","id":1000}`)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	u, err := c.User("john")
	ok(t, err)
	equals(t, 1000, u.ID)
	equals(t, "john", u.Username)
	equals(t, "", u.AccessToken)
}

// Ensure that the GitHub client handles a server error appropriately.
func TestGitHub_User_ErrInternalServerError(t *testing.T) {
	// Create mock GitHub API server that returns an error.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/users/john", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	_, err := c.User("john")
	assert(t, err != nil, "expected error")
}

// Ensure that the GitHub client can retrieve a list of gists by username.
func TestGitHub_Gists(t *testing.T) {
	// Create mock GitHub API server.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/users/foo/gists", r.URL.Path)
		fmt.Fprint(w, `[{"id": "25f126746be9275592eb","html_url": "https://gist.github.com/25f126746be9275592eb","files": {"gistfile1.diff": {"filename": "gistfile1.diff","raw_url": "https://gist.githubusercontent.com/foo/25f126746be9275592eb/raw/ae491a919ac25988dab39677d5784945242f4d04/gistfile1.diff","size": 3073}},"public": true,"description": "My gist","owner": {"login": "foo","id":1000}}]`)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	a, err := c.Gists("foo")
	ok(t, err)

	equals(t, 1, len(a))
	equals(t, "25f126746be9275592eb", a[0].ID)
	equals(t, "foo", a[0].Owner)
	equals(t, "My gist", a[0].Description)
	equals(t, true, a[0].Public)
	equals(t, "https://gist.github.com/25f126746be9275592eb", a[0].URL)

	equals(t, 1, len(a[0].Files))
	equals(t, 3073, a[0].Files[0].Size)
	equals(t, "gistfile1.diff", a[0].Files[0].Filename)
	equals(t, "https://gist.githubusercontent.com/foo/25f126746be9275592eb/raw/ae491a919ac25988dab39677d5784945242f4d04/gistfile1.diff", a[0].Files[0].RawURL)
}

// Ensure that the GitHub client handles a server error appropriately.
func TestGitHub_Gists_ErrInternalServerError(t *testing.T) {
	// Create mock GitHub API server that returns an error.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/users/john/gists", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	_, err := c.Gists("john")
	assert(t, err != nil, "expected error")
}

// Ensure that the GitHub client can retrieve a single gist by ID.
func TestGitHub_Gist(t *testing.T) {
	// Create mock GitHub API server.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/gists/25f126746be9275592eb", r.URL.Path)
		fmt.Fprint(w, `{"id": "25f126746be9275592eb","html_url": "https://gist.github.com/25f126746be9275592eb","files": {"gistfile1.diff": {"filename": "gistfile1.diff","raw_url": "https://gist.githubusercontent.com/foo/25f126746be9275592eb/raw/ae491a919ac25988dab39677d5784945242f4d04/gistfile1.diff","size": 3073}},"public": true,"description": "My gist","owner": {"login": "foo","id":1000}}`)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	gist, err := c.Gist("25f126746be9275592eb")
	ok(t, err)
	equals(t, "25f126746be9275592eb", gist.ID)
}

// Ensure that the GitHub client handles a server error appropriately.
func TestGitHub_Gist_ErrInternalServerError(t *testing.T) {
	// Create mock GitHub API server that returns an error.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		equals(t, "/gists/25f126746be9275592eb", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	// Create client and request the user "john".
	c := gist.NewGitHubClient("xyz")
	c.SetBaseURL(s.URL)
	_, err := c.Gist("25f126746be9275592eb")
	assert(t, err != nil, "expected error")
}

// MockGitHubClient is a mockable GitHub client.
type MockGitHubClient struct {
	UserFunc  func(username string) (*gist.User, error)
	GistsFunc func(username string) ([]*gist.Gist, error)
	GistFunc  func(id string) (*gist.Gist, error)
}

func NewMockGitHubClient(_ string) gist.GitHubClient {
	return &MockGitHubClient{}
}

func (m *MockGitHubClient) SetBaseURL(rawurl string) {}

func (m *MockGitHubClient) User(username string) (*gist.User, error) {
	return m.UserFunc(username)
}

func (m *MockGitHubClient) Gists(username string) ([]*gist.Gist, error) {
	return m.GistsFunc(username)
}

func (m *MockGitHubClient) Gist(id string) (*gist.Gist, error) {
	return m.GistFunc(id)
}
