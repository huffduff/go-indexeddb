package indexeddb

import (
	"fmt"

	"github.com/huffduff/go-indexeddb/indexeddb/internal"
)

type Migration interface{}

type MigrationTransaction struct {
	def *internal.Database
	tr  *Transaction
}

func (p *MigrationTransaction) CreateStore(name string, opts StoreOptions) (*MigrationTransactionStore, error) {
	spec := internal.Store{
		Name:          name,
		KeyPath:       opts.KeyPath,
		AutoIncrement: opts.AutoIncrement,
	}
	h, err := p.def.CreateStore(p.tr.h, spec)
	if err != nil {
		return nil, err
	}

	store := TransactionStore{BaseStore{h}, p.tr}
	return &MigrationTransactionStore{store}, nil
}

type MigrationTransactionStore struct {
	TransactionStore
	// tr MigrationTransaction
}

func (p *MigrationTransactionStore) CreateIndex(name string, opts IndexOptions) error {
	spec := internal.Index{
		Name:       name,
		StoreName:  p.def.Name,
		KeyPath:    opts.KeyPath,
		Unique:     opts.Unique,
		MultiEntry: opts.MultiEntry,
	}
	_, err := p.def.CreateIndex(p.Transaction.h, spec)
	return err
}

func (p *MigrationTransactionStore) DeleteIndex(name string) error {
	// TODO we shouldn't need an instance, just the prefix
	idx, ok := p.def.Indexes[name]
	if !ok {
		return fmt.Errorf("index %s does not exist", name)
	}
	return p.def.DeleteIndex(p.Transaction.h, idx)
}

type migrator struct {
	Migrate func(f func(version uint, h *MigrationTransaction) error) (*Database, error)
}

// migrateError ignores the callback and immediately return the error
func migrateError(err error) *migrator {
	return &migrator{
		func(_ func(_ uint, _ *MigrationTransaction) error) (*Database, error) {
			return nil, err
		},
	}
}

// migrateDone ignores the callback and immediately return the db
func migrateDone(current *internal.Database) *migrator {
	return &migrator{
		func(_ func(_ uint, _ *MigrationTransaction) error) (*Database, error) {
			err := current.Hydrate()
			if err != nil {
				return nil, err
			}
			return &Database{current}, nil
		},
	}
}

// migrateRun returns an updated handle for the database if all migrations succeed
func migrateRun(current *internal.Database, to uint) *migrator {
	return &migrator{
		func(callback func(v uint, h *MigrationTransaction) error) (*Database, error) {
			t, err := newTransaction(current, current.StoreNames(), Default)
			if err != nil {
				return nil, err
			}

			err = callback(current.Version, &MigrationTransaction{current, t})
			if err != nil {
				t.h.Discard()
				return nil, fmt.Errorf("migration discarded %w", err)
			}

			current.Version = to

			err = current.UpdateDefinition(t.h)
			if err == nil {
				err = t.Commit()
			}
			if err != nil {
				t.h.Discard()
				return nil, fmt.Errorf("migration commit failed %w", err)
			}

			err = current.Hydrate()
			if err != nil {
				return nil, err
			}
			return &Database{current}, nil
		},
	}
}
