package gist

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// DB represents the application-level database.
type DB struct {
	*bolt.DB
	secret []byte

	// GistPath to the root of the gist data.
	GistPath string

	// NewGitHubClient is the function used to return a new github client.
	NewGitHubClient func(string) GitHubClient
}

// Open opens and initializes the database.
func (db *DB) Open(path string, mode os.FileMode) error {
	d, err := bolt.Open(path, mode, nil)
	if err != nil {
		return err
	}
	db.DB = d

	if db.NewGitHubClient == nil {
		db.NewGitHubClient = NewGitHubClient
	}

	return db.Update(func(tx *Tx) error {
		// Initialize the top-level buckets.
		_, _ = tx.CreateBucketIfNotExists([]byte("meta"))
		_, _ = tx.CreateBucketIfNotExists([]byte("users"))
		_, _ = tx.CreateBucketIfNotExists([]byte("gists"))

		// Initialize secret.
		if err := tx.GenerateSecretIfNotExists(); err != nil {
			return err
		}
		db.secret = tx.Secret()

		return nil
	})
}

// View executes a function in the context of a read-only transaction.
func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Update executes a function in the context of a writable transaction.
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Secret returns the secure secret key.
func (db *DB) Secret() []byte {
	return db.secret
}

// LoadGist retrieves the latest gist files from GitHub.
func (db *DB) LoadGist(userID int, gistID string) error {
	return db.Update(func(tx *Tx) error {
		// Retrieve user.
		u, err := tx.User(userID)
		if err != nil {
			return fmt.Errorf("user: %s", err)
		} else if u == nil {
			return fmt.Errorf("user not found: %d", userID)
		}

		// Create GitHub client.
		client := db.NewGitHubClient(u.AccessToken)

		// Retrieve gist data.
		gist, err := client.Gist(gistID)
		if err != nil {
			return fmt.Errorf("gist: %s", err)
		} else if gist == nil {
			return fmt.Errorf("gist not found: %s", gistID)
		}

		// Download all files over HTTP.
		ch := make(chan error)
		for _, file := range gist.Files {
			go func(file *GistFile) {
				defer autonotify()
				var err error
				if err = download(file.RawURL, db.GistFilePath(gistID, file.Filename)); err != nil {
					err = fmt.Errorf("download: %s: %s", file.RawURL, err)
				}
				ch <- err
			}(file)
		}

		// Check for download errors.
		for i := 0; i < len(gist.Files); i++ {
			if err := <-ch; err != nil {
				return err
			}
		}

		// Save to the database.
		if err := tx.SaveGist(gist); err != nil {
			return fmt.Errorf("save gist: %s", err)
		}

		return nil
	})
}

// GistFilePath returns the path for a given gist file.
func (db *DB) GistFilePath(gistID, filename string) string {
	return filepath.Join(db.GistPath, gistID, filename)
}

// Tx represents an application-level transaction.
type Tx struct {
	*bolt.Tx
}

func (tx *Tx) meta() *bolt.Bucket  { return tx.Bucket([]byte("meta")) }
func (tx *Tx) gists() *bolt.Bucket { return tx.Bucket([]byte("gists")) }
func (tx *Tx) users() *bolt.Bucket { return tx.Bucket([]byte("users")) }

// Gist retrieves a gist from the database by ID.
func (tx *Tx) Gist(id string) (g *Gist, err error) {
	if v := tx.gists().Get([]byte(id)); v != nil {
		err = json.Unmarshal(v, &g)
	}
	return
}

// SaveGist stores a gist in the database.
func (tx *Tx) SaveGist(g *Gist) error {
	assert(g != nil, "nil gist")
	assert(g.ID != "", "gist id required")
	b, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("marshal gist: %s", err)
	}
	return tx.gists().Put([]byte(g.ID), b)
}

// User retrieves an user from the database by ID.
func (tx *Tx) User(id int) (u *User, err error) {
	if v := tx.users().Get(i64tob(int64(id))); v != nil {
		err = json.Unmarshal(v, &u)
	}
	return
}

// SaveUser stores an user in the database.
func (tx *Tx) SaveUser(u *User) error {
	assert(u != nil, "nil user")
	assert(u.ID != 0, "user id required")
	b, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user: %s", err)
	}
	return tx.users().Put(i64tob(int64(u.ID)), b)
}

// Secret returns the 64-byte secret key.
func (tx *Tx) Secret() []byte {
	return tx.meta().Get([]byte("secret"))
}

// GenerateSecretIfNotExists generates a 64-byte secret key.
func (tx *Tx) GenerateSecretIfNotExists() error {
	// Ignore if a secret is already set.
	if tx.Secret() != nil {
		return nil
	}

	// Otherwise generate one.
	value := make([]byte, 64)
	if _, err := rand.Read(value); err != nil {
		return err
	}
	return tx.meta().Put([]byte("secret"), value)
}

// download retrieves a URL over HTTP GET and writes the response to the path.
func download(url, path string) error {
	// Create the parent directory.
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// Retrieve the file over HTTP.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check the response code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid HTTP status: %d", resp.StatusCode)
	}

	// Open the file to write to.
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create: %s", err)
	}
	defer func() { _ = f.Close() }()

	// Copy from the response to the file.
	// Note: This is not perfect as a partial file can be read but this is
	// not the record of authority so we can always retry if there's a problem.
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

// Converts an integer to a big-endian encoded byte slice.
func i64tob(v int64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
