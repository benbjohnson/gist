package gist

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"
)

// DefaultFilename is the default file used if none is specified in the URL.
const DefaultFilename = "index.html"

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
	case "favicon.ico":
		http.NotFound(w, r)
	default:
		h.gist(w, r)
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

	// Retrieve available gists from GitHub.
	client := NewGitHubClient(user.AccessToken)
	gists, err := client.Gists("")
	if err != nil {
		log.Println("github gists:", err)
		http.Error(w, "github api error", http.StatusInternalServerError)
		return
	}

	// Write gists out.
	(&tmpl{}).Index(w, gists)
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

// gist serves a single file for a gist.
// If the root is requested then the gist content is refreshed.
func (h *Handler) gist(w http.ResponseWriter, r *http.Request) {
	session := h.Session(r)

	// Break up URL path components.
	// If there are less than 3 components, return 404.
	// If we have 3 components, redirect to append a slash.
	a := strings.Split(r.URL.Path, "/")
	if len(a) < 3 || a[0] == "_" {
		log.Println("not found:", r.URL.Path)
		http.NotFound(w, r)
		return
	} else if len(a) == 3 {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	// Extract the values from the URL.
	gistID, filename := a[2], strings.Join(a[3:], "/")

	// Set default filename.
	if filename == "" {
		filename = DefaultFilename
	}

	// Only reload if the following conditions are met:
	//
	//   1. User is logged in.
	//   2. User is loading / or /index.html.
	//   3. User is loading page directly (i.e. not in an iframe).
	//
	reload := (session.Authenticated() && filename == DefaultFilename && r.Referer() == "")

	// Update gist.
	if reload {
		if err := h.db.LoadGist(session.UserID(), gistID); err != nil {
			log.Printf("reload gist: %s", err)
			http.Error(w, "error loading gist", http.StatusInternalServerError)
			return
		}
	}

	// Serve gist file from disk cache.
	path := h.db.GistFilePath(gistID, filename)
	f, err := os.Open(path)
	if err != nil {
		log.Printf("read gist: %s: %s", path, err)
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()

	// Set the content type.
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))

	// Copy the file to the response.
	_, _ = io.Copy(w, f)
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

// tmpl is a namespace for templates
type tmpl struct{}
