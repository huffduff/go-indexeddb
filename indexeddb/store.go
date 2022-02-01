package indexeddb

import (
	"github.com/huffduff/go-indexeddb/indexeddb/internal"
)

// StoreOptions is an optional object with one of two properties:
type StoreOptions struct {

	// keyPath – a path to an object property that IndexedDB will use as the key, e.g. id.
	KeyPath string `json:"keyPath,omitempty"`

	// autoIncrement – if true, then the key for a newly stored object is generated automatically,
	// as an ever-incrementing number.
	AutoIncrement bool `json:"autoIncrement,omitempty"`
}

type Store interface {
	Name() string
	KeyPath() string
	IndexNames() []string
	AutoIncrement() bool
}

type ReadStore interface {
	Store
	GetExact(key Key, val interface{}) error
	Get(query Range, val interface{}) error
	GetMulti(key []Key, val interface{}) error
	GetAll(query Range, limit int, val interface{}) error
	GetKey(query Range) (Key, error)
	GetAllKeys(query Range, limit int) ([]Key, error)
	Count(query Range) (uint, error)
	OpenCursor(query Range, direction Direction) (Cursor, error)
	OpenKeyCursor(query Range, direction Direction) (KeyCursor, error)

	Index(name string) *Index
}

type WriteStore interface {
	ReadStore
	// Put(value interface{}) error
	PutWithKey(key Key, value interface{}) error
	// Add(value interface{}) error
	AddWithKey(key Key, value interface{}) error
	Delete(key Key) error
	Clear() error
}

var _ Store = (*BaseStore)(nil)

var _ Store = (*ReadonlyStore)(nil)
var _ ReadStore = (*ReadonlyStore)(nil)

var _ Store = (*TransactionStore)(nil)
var _ ReadStore = (*TransactionStore)(nil)
var _ WriteStore = (*TransactionStore)(nil)

type BaseStore struct {
	def *internal.Store
	// Transaction *Transaction
}

func (p *BaseStore) Name() string {
	return p.def.Name
}

func (p *BaseStore) KeyPath() string {
	return p.def.KeyPath
}

func (p *BaseStore) IndexNames() []string {
	return p.def.IndexNames()
}

func (p *BaseStore) AutoIncrement() bool {
	return p.def.AutoIncrement
}

type ReadonlyStore struct {
	BaseStore
	Transaction *ReadonlyTransaction
}

func (p *ReadonlyStore) GetExact(key Key, v interface{}) error {
	return p.def.GetExact(p.Transaction.h, key, v)
}

func (p *ReadonlyStore) Get(query Range, v interface{}) error {
	return p.def.Get(p.Transaction.h, query, v)
}

func (p *ReadonlyStore) GetMulti(keys []Key, v interface{}) error {
	return p.def.GetMulti(p.Transaction.h, keys, v)
}

func (p *ReadonlyStore) GetAll(query Range, limit int, v interface{}) error {
	return p.def.GetAll(p.Transaction.h, query, limit, v)
}

func (p *ReadonlyStore) GetKey(query Range) (Key, error) {
	return p.def.GetKey(p.Transaction.h, query)
}

func (p *ReadonlyStore) GetAllKeys(query Range, limit int) ([]Key, error) {
	return p.def.GetAllKeys(p.Transaction.h, query, limit)
}

func (p *ReadonlyStore) Count(query Range) (uint, error) {
	return p.def.Count(p.Transaction.h, query)
}

func (p *ReadonlyStore) OpenCursor(query Range, direction Direction) (Cursor, error) {
	return p.def.GetCursor(p.Transaction.h, query, direction)
}

func (p *ReadonlyStore) OpenKeyCursor(query Range, direction Direction) (KeyCursor, error) {
	return p.def.GetCursor(p.Transaction.h, query, direction)
}

func (p *ReadonlyStore) Index(name string) *Index {
	idx := p.def.Indexes[name]
	return &Index{idx, p, p.Transaction.h}
}

type TransactionStore struct {
	BaseStore
	Transaction *Transaction
}

func (p *TransactionStore) PutWithKey(key Key, value interface{}) error {
	return p.def.Put(p.Transaction.h, key, value)
}

func (p *TransactionStore) AddWithKey(key Key, value interface{}) error {
	return p.def.Add(p.Transaction.h, key, value)
}

func (p *TransactionStore) Delete(key Key) error {
	return p.def.Delete(*p.Transaction.h, key)
}

func (p *TransactionStore) Clear() error {
	return p.def.Clear(*p.Transaction.h)
}

func (p *TransactionStore) GetExact(key Key, v interface{}) error {
	return p.def.GetExact(p.Transaction.h, key, v)
}

func (p *TransactionStore) Get(query Range, v interface{}) error {
	return p.def.Get(p.Transaction.h, query, v)
}

func (p *TransactionStore) GetMulti(keys []Key, v interface{}) error {
	return p.def.GetMulti(p.Transaction.h, keys, v)
}

func (p *TransactionStore) GetAll(query Range, limit int, v interface{}) error {
	return p.def.GetAll(p.Transaction.h, query, limit, v)
}

func (p *TransactionStore) GetKey(query Range) (Key, error) {
	return p.def.GetKey(p.Transaction.h, query)
}

func (p *TransactionStore) GetAllKeys(query Range, limit int) ([]Key, error) {
	return p.def.GetAllKeys(p.Transaction.h, query, limit)
}

func (p *TransactionStore) Count(query Range) (uint, error) {
	return p.def.Count(p.Transaction.h, query)
}

func (p *TransactionStore) OpenCursor(query Range, direction Direction) (Cursor, error) {
	return p.def.GetCursor(p.Transaction.h, query, direction)
}

func (p *TransactionStore) OpenKeyCursor(query Range, direction Direction) (KeyCursor, error) {
	return p.def.GetCursor(p.Transaction.h, query, direction)
}

func (p *TransactionStore) Index(name string) *Index {
	idx := p.def.Indexes[name]
	return &Index{idx, p, p.Transaction.h}
}
