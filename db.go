package gist

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	d, err := bolt.Open(path, mode, nil)
	if err != nil {
		return err
	}
	db.DB = d

	return db.Update(func(tx *Tx) error {
		// Initialize the top-level buckets.
		var meta, _ = tx.CreateBucketIfNotExists([]byte("meta"))
		_, _ = tx.CreateBucketIfNotExists([]byte("users"))
		_, _ = tx.CreateBucketIfNotExists([]byte("gists"))

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
		return fn(&Tx{tx})
	})
}

// Update executes a function in the context of a writable transaction.
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Tx represents an application-level transaction.
type Tx struct {
	*bolt.Tx
}

// users retrieves the users bucket.
func (tx *Tx) users() *bolt.Bucket { return tx.Bucket([]byte("users")) }

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

// Converts an integer to a big-endian encoded byte slice.
func i64tob(v int64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
