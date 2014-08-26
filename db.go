package gist

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

// DB represents the application-level database.
type DB struct {
	*bolt.DB
	secret []byte
}

// Open opens and initializes the database.
func (db *DB) Open(path string, mode os.FileMode) error {
	d, err := bolt.Open(path, mode, nil)
	if err != nil {
		return err
	}
	db.DB = d

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
func (tx *Tx) User(id int64) (u *User, err error) {
	if v := tx.users().Get(i64tob(id)); v != nil {
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

// Converts an integer to a big-endian encoded byte slice.
func i64tob(v int64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
