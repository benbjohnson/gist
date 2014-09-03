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
