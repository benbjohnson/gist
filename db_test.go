package gist_test

import (
	"log"
	"os"
	"testing"

	"github.com/benbjohnson/gist"
)

// Ensure that a database can be opened and closed.
func TestDB_Open(t *testing.T) {
	db := NewTestDB()
	ok(t, db.Close())
}

// Ensure that a gist can be persisted to the database.
func TestTx_SaveGist(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	data := &gist.Gist{
		ID:       "xxx",
		Username: "john",
		CTime:    parsetime("2000-01-01T00:00:00Z"),
		MTime:    parsetime("2010-01-01T00:00:00Z"),
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
