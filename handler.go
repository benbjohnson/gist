package gist

import (
	"net/http"

	//"github.com/boltdb/bolt"
)

// Handler represents the root HTTP handler for the application.
type Handler struct {
	DB     *DB
	Path   string
	Token  string
	Secret string
}

// ServeHTTP dispatches incoming HTTP requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.root(w, r)
	case "/_/authorize/":
		h.authorize(w, r)
	case "/_/authorized/":
		h.authorized(w, r)
	}
}

// root serves the home page.
func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	// TODO(benbjohnson)
}

// authorize redirects the user to GitHub OAuth2 authorization.
func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	// TODO(benbjohnson)
}

// authorized receives the GitHub OAuth2 callback.
func (h *Handler) authorized(w http.ResponseWriter, r *http.Request) {
	// TODO(benbjohnson)
}
