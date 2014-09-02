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
		cert    = flag.String("cert", "", "SSL certificate file")
		key     = flag.String("key", "", "SSL key file")
	)
	flag.Parse()
	log.SetFlags(0)

	// Validate flags.
	if *datadir == "" {
		log.Fatal("data directory required: -d PATH")
	} else if *token == "" {
		log.Fatal("GitHub API token required: -token TOKEN")
	} else if *secret == "" {
		log.Fatal("GitHub API secret required: -secret SECRET")
	} else if *cert != "" && *key == "" {
		log.Fatal("key file required: -key PATH")
	} else if *key != "" && *cert == "" {
		log.Fatal("certificate file required: -cert PATH")
	}

	// Make sure the data directory exists.
	if err := os.MkdirAll(*datadir, 0700); err != nil {
		log.Fatal(err)
	}

	// Open the database.
	var db gist.DB
	db.GistPath = filepath.Join(*datadir, "gists")
	if err := db.Open(filepath.Join(*datadir, "db"), 0600); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	// Initialize the handler.
	h := gist.NewHandler(&db, *datadir, *token, *secret)

	// Start HTTP server.
	if *cert != "" && *key != "" {
		go func() { log.Fatal(http.ListenAndServeTLS(":443", *cert, *key, h)) }()
	}
	go func() { log.Fatal(http.ListenAndServe(*addr, h)) }()

	log.Printf("Listening on http://localhost%s", *addr)
	log.SetFlags(log.LstdFlags)

	<-(chan struct{})(nil)
}
