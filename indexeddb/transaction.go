package indexeddb

import (
	"fmt"

	"github.com/huffduff/go-indexeddb/indexeddb/internal"
	"github.com/syndtr/goleveldb/leveldb"
)

type TransactionDurability string

const (
	Default TransactionDurability = "default"
	Strict  TransactionDurability = "strict"
	Relaxed TransactionDurability = "relaxed"
)

type baseTransaction struct {
	Durability TransactionDurability

	stores map[string]Store
}

func (p *baseTransaction) StoreNames() []string {
	out := make([]string, 0, len(p.stores))
	for s := range p.stores {
		out = append(out, s)
	}
	return out
}

type ReadonlyTransaction struct {
	baseTransaction
	h *leveldb.Snapshot
}

func (p *ReadonlyTransaction) Store(name string) *ReadonlyStore {
	// what happens if the store isn't in the state?
	return p.stores[name].(*ReadonlyStore)
}

func (p *ReadonlyTransaction) Abort() {
	p.h.Release()
}

func (p *ReadonlyTransaction) Commit() {
	p.h.Release()
}

type Transaction struct {
	baseTransaction
	h *leveldb.Transaction
}

func (p *Transaction) Store(name string) *TransactionStore {
	// what happens if the store isn't in the state?
	return p.stores[name].(*TransactionStore)
}

func (p *Transaction) Abort() {
	p.h.Discard()
}

func (p *Transaction) Commit() error {
	return p.h.Commit()
}

func newReadonlyTransaction(db *internal.Database, scope []string, durability TransactionDurability) (*ReadonlyTransaction, error) {
	h, err := db.GetSnapshot()
	if err != nil {
		return nil, err
	}
	t := &ReadonlyTransaction{
		baseTransaction{
			durability,
			make(map[string]Store, len(scope)),
		},
		h,
	}
	for _, row := range scope {
		store, ok := db.Stores[row]
		if !ok {
			return nil, fmt.Errorf("store %s not found", row)
		}
		t.stores[row] = &ReadonlyStore{BaseStore{store}, t}
	}
	return t, nil
}

func newTransaction(db *internal.Database, scope []string, durability TransactionDurability) (*Transaction, error) {
	h, err := db.OpenTransaction()
	if err != nil {
		return nil, err
	}
	t := &Transaction{
		baseTransaction{
			durability,
			make(map[string]Store, len(scope)),
		},
		h,
	}
	for _, row := range scope {
		store, ok := db.Stores[row]
		if !ok {
			return nil, fmt.Errorf("store %s not found", row)
		}
		t.stores[row] = &TransactionStore{BaseStore{store}, t}
	}
	return t, nil
}
