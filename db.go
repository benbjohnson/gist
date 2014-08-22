package gist

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// DB represents the application-level database.
type DB struct {
	*bolt.DB
	store sessions.Store
}

// Open opens and initializes the database.
func (db *DB) Open(path string, mode os.FileMode) error {
	d, err := bolt.Open(path, mode)
	if err != nil {
		return err
	}
	db.DB = d

	return db.Update(func(tx *Tx) error {
		// Initialize the top-level buckets.
		var meta, _ = tx.CreateBucketIfNotExists([]byte("meta"))
		tx.CreateBucketIfNotExists([]byte("users"))

		// Initialize secure cookie store.
		secret := meta.Get([]byte("secret"))
		if secret == nil {
			secret = securecookie.GenerateRandomKey(64)
			if err := meta.Put([]byte("secret"), secret); err != nil {
				return fmt.Errorf("secret: %s", err)
			}
		}

		// Create the cookie store.
		db.store = sessions.NewCookieStore(secret)

		return nil
	})
}

// View executes a function in the context of a read-only transaction.
func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		return fn(&Tx{tx, db})
	})
}

// Update executes a function in the context of a writable transaction.
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx, db})
	})
}

// Tx represents an application-level transaction.
type Tx struct {
	*bolt.Tx
}
