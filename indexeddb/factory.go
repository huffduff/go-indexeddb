package indexeddb

import (
	"bytes"
	"fmt"

	"github.com/huffduff/go-indexeddb/indexeddb/internal"
)

// Open initializes a database and returns an initialization struct.
// To get a database handle you must call the Migrate method of the returned value.
// The Migrate function will only fire if the exisitng version is lower than the
// requested version and no other database related errors have been triggered.
// returning an error will rollback the migration and fail.
func Open(name string, version uint, path string) *migrator {
	def, err := internal.OpenDatabase(name, path)
	if err != nil {
		return migrateError(err)
	}

	if def.Version > version {
		err = fmt.Errorf("existing database version %d > %d", def.Version, version)
		return migrateError(err)
	}

	if def.Version < version {
		return migrateRun(def, version)
	}
	return migrateDone(def)
}

func DeleteDatabase(name string) error {
	return fmt.Errorf("not implemented")
}

func Cmp(a []byte, b []byte) int {
	return bytes.Compare(a, b)
}

func Databases() (map[string]uint, error) {
	// will likely require a base path
	return nil, fmt.Errorf("not implemented")
}
