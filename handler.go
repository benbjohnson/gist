package gist

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// Handler represents the root HTTP handler for the application.
type Handler struct {
	db     *DB
	path   string
	token  string
	secret string
	store  sessions.Store
}

// NewHandler returns a new instance of Handler.
func NewHandler(db *DB, path, token, secret string) *Handler {
	return &Handler{
		db:     db,
		path:   path,
		token:  token,
		secret: secret,
		store:  sessions.NewCookieStore(db.secret),
	}
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

// Session returns the current session.
func (h *Handler) Session(r *http.Request) *Session {
	s, _ := h.store.Get(r, "default")
	return &Session{s}
}

// root serves the home page.
func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	s := h.Session(r)
	if !s.Authenticated() {
		http.Redirect(w, r, "/_/authorize/", http.StatusFound)
		return
	}

	fmt.Fprintln(w, "ROOT")
}

// authorize redirects the user to GitHub OAuth2 authorization.
func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	// TODO(benbjohnson)
	fmt.Fprintln(w, "AUTHORIZE")
}

// authorized receives the GitHub OAuth2 callback.
func (h *Handler) authorized(w http.ResponseWriter, r *http.Request) {
	// TODO(benbjohnson)
}

// Session represents an HTTP session.
type Session struct {
	*sessions.Session
}

// Authenticated returns true if there is a user attached to the session.
func (s *Session) Authenticated() bool {
	return s.UserID() != 0
}

// UserID returns the user id on the session.
func (s *Session) UserID() int {
	id, _ := s.Values["UserID"].(int)
	return id
}
