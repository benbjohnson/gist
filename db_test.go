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
