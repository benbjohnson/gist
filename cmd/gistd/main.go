package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/benbjohnson/gist"
)

func main() {
	var (
		datadir = flag.String("d", "", "data directory")
		addr    = flag.String("addr", ":40000", "bind address")
		token   = flag.String("token", "", "api token")
		secret  = flag.String("secret", "", "api secret")
	)
	flag.Parse()

	// Validate flags.
	if *datadir == "" {
		log.Fatal("data directory required: -d PATH")
	} else if *token == "" {
		log.Fatal("GitHub API token required: -token TOKEN")
	} else if *secret == "" {
		log.Fatal("GitHub API secret required: -secret SECRET")
	}

	// Make sure the data directory exists.
	if err := os.MkdirAll(*datadir, 0700); err != nil {
		log.Fatal(err)
	}

	// Open the database.
	var db gist.DB
	if err := db.Open(filepath.Join(*datadir, "db"), 0600); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	// Initialize the handler.
	h := &gist.Handler{
		DB:     &db,
		Path:   *datadir,
		Token:  *token,
		Secret: *secret,
	}

	// Start HTTP server.
	log.Fatal(http.ListenAndServe(*addr, h))
}
