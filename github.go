package gist

import (
	"fmt"
	"net/url"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// GitHubClient is an interface for abstracting the GitHub API.
type GitHubClient interface {
	SetBaseURL(u string)
	User(username string) (*User, error)
	Gists(username string) ([]*Gist, error)
	Gist(id string) (*Gist, error)
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

// SetBaseURL sets the base URL for testing.
func (c *gitHubClient) SetBaseURL(rawurl string) { c.BaseURL, _ = url.Parse(rawurl) }

// User returns a user by username.
func (c *gitHubClient) User(username string) (*User, error) {
	// Retrieve user from GitHub.
	user, _, err := c.Users.Get(username)
	if err != nil {
		return nil, fmt.Errorf("get user: %s", err)
	}

	// Convert to our application type.
	u := &User{}
	if user.ID != nil {
		u.ID = *user.ID
	}
	if user.Login != nil {
		u.Username = *user.Login
	}
	return u, nil
}

// Gists returns a list of gists for a user.
func (c *gitHubClient) Gists(username string) ([]*Gist, error) {
	// Retrieve gists from GitHub.
	a, _, err := c.Client.Gists.List(username, nil)
	if err != nil {
		return nil, fmt.Errorf("list gists: %s", err)
	}

	// Convert to our application type.
	var gists []*Gist
	for _, item := range a {
		gist := &Gist{}
		gist.deserializeGist(&item, false)
		gists = append(gists, gist)
	}

	return gists, nil
}

// Gist returns a single gist by ID (with content).
func (c *gitHubClient) Gist(id string) (*Gist, error) {
	// Retrieve gist from GitHub.
	item, _, err := c.Client.Gists.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get gist: %s", err)
	}

	// Convert to our application type.
	gist := &Gist{}
	gist.deserializeGist(item, true)
	return gist, nil
}

func (g *Gist) deserializeGist(item *github.Gist, useContent bool) {
	if item.ID != nil {
		g.ID = *item.ID
	}
	if item.Owner != nil && item.Owner.Login != nil {
		g.Owner = *item.Owner.Login
	}
	if item.Description != nil {
		g.Description = *item.Description
	}
	if item.Public != nil {
		g.Public = *item.Public
	}
	if item.HTMLURL != nil {
		g.URL = *item.HTMLURL
	}

	for _, file := range item.Files {
		f := &GistFile{}
		if file.Size != nil {
			f.Size = *file.Size
		}
		if file.Filename != nil {
			f.Filename = *file.Filename
		}
		if file.RawURL != nil {
			f.RawURL = *file.RawURL
		}
		g.Files = append(g.Files, f)
	}
}
