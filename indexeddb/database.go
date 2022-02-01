package indexeddb

import (
	"github.com/huffduff/go-indexeddb/indexeddb/internal"
)

type Database struct {
	def *internal.Database
}

func (p *Database) Name() string {
	return p.def.Name
}

func (p *Database) Version() uint {
	return p.def.Version
}

func (p *Database) StoreNames() []string {
	return p.def.StoreNames()
}

func (p *Database) Transaction(scope []string, durability TransactionDurability) (*Transaction, error) {
	return newTransaction(p.def, scope, durability)
}

func (p *Database) ReadonlyTransaction(scope []string, durability TransactionDurability) (*ReadonlyTransaction, error) {
	return newReadonlyTransaction(p.def, scope, durability)
}
