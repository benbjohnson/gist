package gist_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/benbjohnson/gist"
)

// Ensure that a database can be opened and closed.
func TestDB_Open(t *testing.T) {
	db := NewTestDB()
	ok(t, db.Close())
}

// Ensure that a database will generate a 64-byte secret and save it to the DB.
func TestDB_Secret(t *testing.T) {
	db := NewTestDB()
	defer db.Close()
	equals(t, 64, len(db.Secret()))
}

// Ensure that a gist can be persisted to the database.
func TestTx_SaveGist(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	data := &gist.Gist{
		ID:          "xxx",
		UserID:      1000,
		Description: "My gist",
		Public:      true,
		URL:         "http://gist.github.com/john/xxx",
		CreatedAt:   time.Now().UTC(),
		Files: []*gist.GistFile{
			{Size: 100, Filename: "index.html", RawURL: "http://raw.github.com/john/xxx/index.html"},
		},
	}

	ok(t, db.Update(func(tx *gist.Tx) error {
		ok(t, tx.SaveGist(data))
		return nil
	}))

	ok(t, db.View(func(tx *gist.Tx) error {
		g, _ := tx.Gist("xxx")
		equals(t, data, g)
		return nil
	}))
}

// Ensure that a user can be persisted to the database.
func TestTx_SaveUser(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	ok(t, db.Update(func(tx *gist.Tx) error {
		ok(t, tx.SaveUser(&gist.User{ID: 100, Username: "john", AccessToken: "1234"}))
		return nil
	}))

	ok(t, db.View(func(tx *gist.Tx) error {
		u, _ := tx.User(100)
		equals(t, &gist.User{ID: 100, Username: "john", AccessToken: "1234"}, u)
		return nil
	}))
}

// TestDB wraps the DB to provide helper functions and clean up.
type TestDB struct {
	*gist.DB
}

func NewTestDB() *TestDB {
	db := &TestDB{DB: &gist.DB{}}
	if err := db.Open(tempfile(), 0600); err != nil {
		log.Fatal("open: ", err)
	}
	return db
}

func (db *TestDB) Close() error {
	defer os.RemoveAll(db.Path())
	return db.DB.Close()
}
