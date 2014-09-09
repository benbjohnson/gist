package gist

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"
)

// errNonCanonicalPath is returned when a URL needs a slash at the end.
var errNonCanonicalPath = errors.New("non-canonical path")

const (
	// DefaultFilename is the default file used if none is specified in the URL.
	DefaultFilename = "index.html"

	// DefaultEmbedWidth is the width returned from the oEmbed endpoint.
	DefaultEmbedWidth = 600

	// DefaultEmbedHeight is the height returned from the oEmbed endpoint.
	DefaultEmbedHeight = 300

	// EmbedCacheAge is the number of seconds a consumer should cache an oEmbed.
	EmbedCacheAge = 0
)

// Handler represents the root HTTP handler for the application.
type Handler struct {
	db     *DB
	config *oauth.Config
	Store  sessions.Store
	Logger *log.Logger

	// NewGitHubClient returns a new GitHub client.
	NewGitHubClient func(string) GitHubClient

	// ExchangeFunc processes a returned OAuth2 code into a token.
	// This function is used for testing.
	ExchangeFunc func(string) (*oauth.Token, error)
}

// NewHandler returns a new instance of Handler.
func NewHandler(db *DB, token, secret string) *Handler {
	h := &Handler{
		db: db,
		config: &oauth.Config{
			ClientId:     token,
			ClientSecret: secret,
			Scope:        "",
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
		},
		Store:           sessions.NewCookieStore(db.secret),
		NewGitHubClient: NewGitHubClient,
		Logger:          log.New(os.Stderr, "", log.LstdFlags),
	}
	h.ExchangeFunc = h.exchangeFunc
	return h
}

// DB returns the database reference.
func (h *Handler) DB() *DB { return h.db }

// ServeHTTP dispatches incoming HTTP requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ignore the noise.
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}

	// Record the start time.
	t := time.Now()

	// Route to the appropriate handlers.
	switch r.URL.Path {
	case "/":
		h.HandleRoot(w, r)
	case "/_/dashboard":
		h.HandleDashboard(w, r)
	case "/_/login":
		h.HandleLogin(w, r)
	case "/_/login/callback":
		h.HandleLoginCallback(w, r)
	case "/_/logout":
		h.HandleLogout(w, r)
	case "/oembed", "/oembed/", "/oembed.xml":
		h.HandleOEmbed(w, r)
	case "/oembed.json":
		h.HandleOEmbedJSON(w, r)
	default:
		h.HandleGist(w, r)
	}

	// Write to access log.
	h.log(r, &t)
}

// log records the HTTP access to stdout.
func (h *Handler) log(r *http.Request, t *time.Time) {
	h.Logger.Printf(`%s %s %q %q`+"\n", r.Method, r.RequestURI, r.Referer(), r.UserAgent())
}

// Session returns the current session.
func (h *Handler) Session(r *http.Request) *Session {
	s, _ := h.Store.Get(r, "default")
	return &Session{s}
}

// HandleRoot serves the home page.
func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Retrieve session. If authorized then forward to the dashboard.
	session := h.Session(r)
	if session.Authenticated() {
		http.Redirect(w, r, "/_/dashboard", http.StatusFound)
		return
	}

	// Render home page.
	_ = (&tmpl{}).Index(w)
}

// HandleDashboard serves the dashboard page.
func (h *Handler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Retrieve session. If not authorized then send to home page.
	session := h.Session(r)
	if !session.Authenticated() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Retrieve user and hosted gists.
	var user *User
	var hosted []*Gist
	err := h.db.View(func(tx *Tx) (err error) {
		if user, err = tx.User(session.UserID()); err != nil {
			return
		}
		if hosted, err = tx.GistsByUserID(session.UserID()); err != nil {
			return
		}
		return
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Retrieve available gists from GitHub.
	client := h.NewGitHubClient(user.AccessToken)
	recent, err := client.Gists("")
	if err != nil {
		h.Logger.Println("github gists:", err)
		http.Error(w, "github api error", http.StatusInternalServerError)
		return
	}

	// Write gists out.
	_ = (&tmpl{}).Dashboard(w, hosted, recent)
}

// HandleLogin redirects the user to GitHub OAuth2 authorization.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate auth state.
	var b [16]byte
	_, _ = rand.Read(b[:])
	state := fmt.Sprintf("%x", b)

	// Save state to session.
	session := h.Session(r)
	session.Values["AuthState"] = state
	_ = session.Save(r, w)

	// Redirect user to GitHub for OAuth authorization.
	http.Redirect(w, r, h.config.AuthCodeURL(state), http.StatusFound)
}

// HandleLoginCallback receives the GitHub OAuth2 callback.
func (h *Handler) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	session := h.Session(r)
	state, _ := session.Values["AuthState"].(string)

	// Verify that the auth code was not tampered with.
	if s := r.FormValue("state"); s != state {
		h.Logger.Printf("tampered state: %q != %q", s, state)
		http.Error(w, "auth state mismatch", http.StatusBadRequest)
		return
	}

	// Extract the access token.
	token, err := h.exchange(r.FormValue("code"))
	if err != nil {
		h.Logger.Println("exchange:", err)
		http.Error(w, "oauth exchange error", http.StatusBadRequest)
		return
	}

	// Retrieve user data.
	client := h.NewGitHubClient(token.AccessToken)
	user, err := client.User("")
	if err != nil {
		h.Logger.Println("github:", err)
		http.Error(w, "github api error", http.StatusInternalServerError)
		return
	}
	user.AccessToken = token.AccessToken

	// Persist user to the database.
	err = h.db.Update(func(tx *Tx) error {
		return tx.SaveUser(user)
	})
	if err != nil {
		h.Logger.Println("save user:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save user id to the session.
	session.Values["UserID"] = user.ID
	_ = session.Save(r, w)

	// Redirect to dashboard page.
	http.Redirect(w, r, "/_/dashboard", http.StatusFound)
}

// HandleLogout removes user authentication.
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Save state to session.
	session := h.Session(r)
	session.Values = make(map[interface{}]interface{})
	_ = session.Save(r, w)

	// Redirect user to home page.
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleOEmbed provides an oEmbed endpoint.
func (h *Handler) HandleOEmbed(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("format") {
	case "json":
		h.HandleOEmbedJSON(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// HandleOEmbedJSON provides an oEmbed endpoint.
func (h *Handler) HandleOEmbedJSON(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Version      string `json:"version"`
		Type         string `json:"type"`
		HTML         string `json:"html"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Title        string `json:"title"`
		CacheAge     int    `json:"cache_age"`
		ProviderName string `json:"provider_name"`
		ProviderURL  string `json:"provider_url"`
	}

	// Retrieve URL parameter and parse.
	u, err := url.Parse(r.FormValue("url"))
	if err != nil {
		h.Logger.Printf("oembed: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	q := u.Query()

	// Retrieve width & height.
	width, height := DefaultEmbedWidth, DefaultEmbedHeight
	if v, _ := strconv.Atoi(q.Get("width")); v > 0 {
		width = v
	}
	if v, _ := strconv.Atoi(q.Get("height")); v > 0 {
		height = v
	}

	// Extract gist id.
	gistID, _, err := ParsePath(u.Path)
	if err == errNonCanonicalPath {
		u.Path += "/"
	} else if err != nil {
		h.Logger.Printf("oembed: parse path: %s", err)
		http.NotFound(w, r)
		return
	}

	// Retrieve gist.
	var gist *Gist
	err = h.db.View(func(tx *Tx) (err error) {
		gist, err = tx.Gist(gistID)
		return nil
	})
	if err != nil {
		h.Logger.Printf("oembed: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	} else if gist == nil {
		h.Logger.Printf("oembed: not found: %s", gistID)
		http.NotFound(w, r)
		return
	}

	// Construct an oEmbed response.
	resp := &response{
		Version:      "1.0",
		Type:         "rich",
		Width:        width,
		Height:       height,
		Title:        gist.Description,
		CacheAge:     EmbedCacheAge,
		ProviderName: "Gist Exposed!",
		ProviderURL:  "https://gist.exposed",
	}

	// Write out the JSON-encoded response.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Println("json:", err)
	}
}

// HandleGist serves a single file for a gist.
// If the root is requested then the gist content is refreshed.
func (h *Handler) HandleGist(w http.ResponseWriter, r *http.Request) {
	session := h.Session(r)

	// Extract the path variables.
	gistID, filename, err := ParsePath(r.URL.Path)
	if err == errNonCanonicalPath {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	} else if err != nil {
		h.Logger.Println("parse path:", err)
		http.NotFound(w, r)
		return
	}

	// Set default filename.
	if filename == "" {
		filename = DefaultFilename
	}

	// Parse referrer.
	referrer, _ := url.Parse(r.Referer())

	// Only reload if the following conditions are met:
	//
	//   1. User is logged in.
	//   2. User is loading an HTML page.
	//   3. User is loading page directly (i.e. not in an iframe).
	//
	reload := true
	reload = reload && session.Authenticated()
	reload = reload && filepath.Ext(filename) == ".html"
	reload = reload && (r.Referer() == "" || referrer.Host == r.Host)

	// Update gist.
	if reload {
		if err := h.db.LoadGist(session.UserID(), gistID); err != nil {
			h.Logger.Printf("reload gist: %s", err)
			http.Error(w, "error loading gist", http.StatusInternalServerError)
			return
		}
	}

	// Serve gist file from disk cache.
	path := h.db.GistFilePath(gistID, filename)
	f, err := os.Open(path)
	if err != nil {
		h.Logger.Printf("read gist: %s: %s", path, err)
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()

	// Set the content type.
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(filename)))

	// Copy the file to the response.
	_, _ = io.Copy(w, f)
}

func (h *Handler) exchange(code string) (*oauth.Token, error) {
	return h.ExchangeFunc(code)
}

func (h *Handler) exchangeFunc(code string) (*oauth.Token, error) {
	var t = &oauth.Transport{Config: h.config}
	return t.Exchange(code)
}

// ParsePath extracts the gist id and filename from the path.
func ParsePath(s string) (gistID, filename string, err error) {
	a := strings.Split(s, "/")[1:]
	switch len(a) {
	case 1:
		if a[0] == "" {
			return "", "", fmt.Errorf("invalid path")
		}
		return a[0], "", errNonCanonicalPath
	case 2:
		if strings.Contains(a[1], ".") || a[1] == "" {
			return a[0], a[1], nil
		}
		return a[1], "", errNonCanonicalPath
	case 3:
		return a[1], a[2], nil
	default:
		return "", "", fmt.Errorf("invalid path: %s", s)
	}
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
