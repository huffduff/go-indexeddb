package indexeddb

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Database struct {
	*leveldb.DB
}

func (p *Database) OpenFile(name string) (*Database, error) {
	db, err := leveldb.OpenFile(name, nil)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}
