package gist

import (
	"fmt"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// GitHubClient is an interface for abstracting the GitHub API.
type GitHubClient interface {
	User(username string) (*User, error)
	Gists(username string) ([]*Gist, error)
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
		gist.deserializeGist(&item)
		gists = append(gists, gist)
	}

	return gists, nil
}

func (g *Gist) deserializeGist(item *github.Gist) {
	g.ID = *item.ID
	g.Owner = *item.Owner.Login
	g.Description = *item.Description
	g.Public = *item.Public
	g.URL = *item.HTMLURL

	for _, file := range item.Files {
		f := &GistFile{
			Size:     *file.Size,
			Filename: *file.Filename,
			RawURL:   *file.RawURL,
		}

		if file.Content != nil {
			f.Content = []byte(*file.Content)
		}

		g.Files = append(g.Files, f)
	}
}
