package store

import (
	"database/sql"
)

// A new database to store information
// for the maintainer
type Datastore struct {
	db *sql.DB
}

// Return a new Datastore using the passed database
func NewDatastore(db *sql.DB) (*Datastore, error) {
	return &Datastore{db}, nil
}

func (ds *Datastore) Close() error {
	return ds.db.Close()
}

func (ds *Datastore) AddRepository(org, name string) error {
	_, err := ds.db.Exec("INSERT INTO repository (org, name) VALUES (?,?);", org, name)
	return err
}
