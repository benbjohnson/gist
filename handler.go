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
	path   string
	config *oauth.Config
	store  sessions.Store

	// NewGitHubClient returns a new GitHub client.
	NewGitHubClient func(string) GitHubClient
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
		NewGitHubClient: NewGitHubClient,
	}
}

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
		h.root(w, r)
	case "/_/authorize":
		h.authorize(w, r)
	case "/_/authorized":
		h.authorized(w, r)
	case "/oembed", "/oembed/":
		h.oembed(w, r)
	case "/oembed.json":
		h.oembedJSON(w, r)
	case "/oembed.xml":
		h.oembedXML(w, r)
	default:
		h.gist(w, r)
	}

	// Write to access log.
	h.log(r, &t)
}

// log records the HTTP access to stdout.
func (h *Handler) log(r *http.Request, t *time.Time) {
	fmt.Printf(`%s - - [%s] "%s %s %s" - - %q %q`+"\n", r.RemoteAddr, t.Format("02/Jan/2006:15:04:05 -0700"), r.Method, r.RequestURI, r.Proto, r.Referer(), r.UserAgent())
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
	client := h.NewGitHubClient(user.AccessToken)
	gists, err := client.Gists("")
	if err != nil {
		log.Println("github gists:", err)
		http.Error(w, "github api error", http.StatusInternalServerError)
		return
	}

	// Write gists out.
	_ = (&tmpl{}).Index(w, gists)
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
	_ = session.Save(r, w)

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
	_ = session.Save(r, w)

	// Redirect to home page.
	http.Redirect(w, r, "/", http.StatusFound)
}

// oembed provides an oEmbed endpoint.
func (h *Handler) oembed(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("format") {
	case "json":
		h.oembedJSON(w, r)
	case "xml":
		h.oembedXML(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// oembed provides an oEmbed endpoint.
func (h *Handler) oembedJSON(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Version      string `json:"version"`
		Type         string `json:"type"`
		HTML         string `json:"html"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Title        string `json:"title"`
		CacheAge     int    `json:"cache_age"`
		AuthorName   string `json:"author_name"`
		AuthorURL    string `json:"author_url"`
		ProviderName string `json:"provider_name"`
		ProviderURL  string `json:"provider_url"`
	}

	// Retrieve URL parameter and parse.
	u, err := url.Parse(r.FormValue("url"))
	if err != nil {
		log.Printf("oembed: %s", err)
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
		log.Printf("oembed: parse path: %s", err)
		http.NotFound(w, r)
	}

	// Retrieve gist.
	var gist *Gist
	err = h.db.View(func(tx *Tx) (err error) {
		gist, err = tx.Gist(gistID)
		return nil
	})
	if err != nil {
		log.Printf("oembed: %s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	} else if gist == nil {
		log.Printf("oembed: not found: %s", gistID)
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
		AuthorName:   gist.Owner,
		AuthorURL:    (&url.URL{Scheme: "https", Host: "github.com", Path: "/" + gist.Owner}).String(),
		ProviderName: "Gist Exposed!",
		ProviderURL:  "https://gist.exposed",
	}

	// Write out the JSON-encoded response.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("json:", err)
	}
}

// oembedXML provides an oEmbed XML endpoint.
func (h *Handler) oembedXML(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// gist serves a single file for a gist.
// If the root is requested then the gist content is refreshed.
func (h *Handler) gist(w http.ResponseWriter, r *http.Request) {
	session := h.Session(r)

	// Extract the path variables.
	gistID, filename, err := ParsePath(r.URL.Path)
	if err == errNonCanonicalPath {
		u := r.URL
		u.Path += "/"
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	} else if err != nil {
		log.Println("parse path:", err)
		http.NotFound(w, r)
		return
	}

	// Set default filename.
	if filename == "" {
		filename = DefaultFilename
	}

	// Only reload if the following conditions are met:
	//
	//   1. User is logged in.
	//   2. User is loading an HTML page.
	//   3. User is loading page directly (i.e. not in an iframe).
	//
	reload := (session.Authenticated() && filepath.Ext(filename) == ".html" && r.Referer() == "")

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

// ParsePath extracts the gist id and filename from the path.
func ParsePath(s string) (gistID, filename string, err error) {
	// Break apart path into components.
	// A path is invalid if we have less than 3 components or it's a reserved path.
	// If we have exactly 3 components, it's non-canonical and should be redirected.
	a := strings.Split(s, "/")
	if len(a) < 3 || a[0] == "_" {
		err = fmt.Errorf("invalid path: %s", s)
		return
	} else if len(a) == 3 {
		err = errNonCanonicalPath
		return
	}

	// Extract the values from the URL.
	gistID, filename = a[2], strings.Join(a[3:], "/")
	return
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
