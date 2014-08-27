package gist

import (
	"fmt"
	"os"
)

// Gist represents a single GitHub gist.
type Gist struct {
	ID          string      `json:"id"`
	Owner       string      `json:"owner"`
	Description string      `json:"description"`
	Public      bool        `json:"public"`
	URL         string      `json:"url"`
	Files       []*GistFile `json:"files"`
}

// GistFile represents an individual file within a gist.
type GistFile struct {
	Size     int    `json:"size"`
	Filename string `json:"filename"`
	RawURL   string `json:"rawURL"`
	Content  []byte `json:"-"`
}

// User represents a GitHub authorized user on the system.
type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	AccessToken string `json:"accessToken"`
}

// assert will panic with a formatted message if the condition is false.
func assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assert failed: "+msg, v...))
	}
}

func warn(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

func warnf(msg string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", v...)
}
