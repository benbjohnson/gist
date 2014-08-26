package gist

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"
)

// Handler represents the root HTTP handler for the application.
type Handler struct {
	db     *DB
	path   string
	config *oauth.Config
	store  sessions.Store
}

// NewHandler returns a new instance of Handler.
func NewHandler(db *DB, path, token, secret string) *Handler {
	return &Handler{
		db:    db,
		path:  path,
		store: sessions.NewCookieStore(db.secret),
		config: &oauth.Config{
			ClientId:     token,
			ClientSecret: secret,
			Scope:        "",
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
		},
	}
}

// ServeHTTP dispatches incoming HTTP requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.root(w, r)
	case "/_/authorize":
		h.authorize(w, r)
	case "/_/authorized":
		h.authorized(w, r)
	default:
		log.Println("not found:", r.URL.Path)
		http.NotFound(w, r)
	}
}

// Session returns the current session.
func (h *Handler) Session(r *http.Request) *Session {
	s, _ := h.store.Get(r, "default")
	return &Session{s}
}

// root serves the home page.
func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	// Retrieve session. If not authorized then send to GitHub.
	session := h.Session(r)
	if !session.Authenticated() {
		http.Redirect(w, r, "/_/authorize", http.StatusFound)
		return
	}

	// Retrieve user.
	var user *User
	err := h.db.View(func(tx *Tx) (err error) {
		user, err = tx.User(session.UserID())
		return
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// TODO(benbjohnson): Show a list of hosted gists.
	// TODO(benbjohnson): Show a list of available gists.

	// TEMP: Print out user data.
	json.NewEncoder(w).Encode(user)
}

// authorize redirects the user to GitHub OAuth2 authorization.
func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	// Generate auth state.
	var b [16]byte
	_, _ = rand.Read(b[:])
	state := fmt.Sprintf("%x", b)

	// Save state to session.
	session := h.Session(r)
	session.Values["AuthState"] = state
	session.Save(r, w)

	// Redirect user to GitHub for OAuth authorization.
	http.Redirect(w, r, h.config.AuthCodeURL(state), http.StatusFound)
}

// authorized receives the GitHub OAuth2 callback.
func (h *Handler) authorized(w http.ResponseWriter, r *http.Request) {
	session := h.Session(r)
	state, _ := session.Values["AuthState"].(string)

	// Verify that the auth code was not tampered with.
	if s := r.FormValue("state"); s != state {
		log.Printf("tampered state: %q != %q", s, state)
		http.Error(w, "auth state mismatch", http.StatusBadRequest)
		return
	}

	// Extract the access token.
	var t = &oauth.Transport{Config: h.config}
	token, err := t.Exchange(r.FormValue("code"))
	if err != nil {
		log.Println("exchange:", err)
		http.Error(w, "oauth exchange error", http.StatusBadRequest)
		return
	}

	// Retrieve user data.
	client := NewGitHubClient(token.AccessToken)
	user, err := client.User("")
	if err != nil {
		log.Println("github:", err)
		http.Error(w, "github api error", http.StatusInternalServerError)
		return
	}
	user.AccessToken = token.AccessToken

	// Persist user to the database.
	err = h.db.Update(func(tx *Tx) error {
		return tx.SaveUser(user)
	})
	if err != nil {
		log.Println("save user:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save user id to the session.
	session.Values["UserID"] = user.ID
	session.Save(r, w)

	// Redirect to home page.
	http.Redirect(w, r, "/", http.StatusFound)
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
