package gist

import (
	"fmt"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// GitHubClient is an interface for abstracting the GitHub API.
type GitHubClient interface {
	User(username string) (*User, error)
}

// NewGitHubClient returns an instance of GitHubClient using a given access token.
func NewGitHubClient(token string) GitHubClient {
	t := &oauth.Transport{Token: &oauth.Token{AccessToken: token}}
	return &gitHubClient{github.NewClient(t.Client())}
}

// gitHubClient wraps the third-party client to implement the GitHubClient interface.
type gitHubClient struct {
	*github.Client
}

// User returns a user by username.
func (c *gitHubClient) User(username string) (*User, error) {
	// Retrieve user from GitHub.
	user, _, err := c.Users.Get("")
	if err != nil {
		return nil, fmt.Errorf("get user: %s", err)
	}

	// Convert to our application type.
	u := &User{
		ID:       *user.ID,
		Username: *user.Login,
		// TODO: AccessToken?
	}
	return u, nil
}
