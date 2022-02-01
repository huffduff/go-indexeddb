package indexeddb

import (
	"github.com/huffduff/go-indexeddb/indexeddb/internal"
	"github.com/syndtr/goleveldb/leveldb"
)

type Cursor = internal.Cursor
type KeyCursor = internal.KeyCursor
type MultiKeyCursor = internal.MultiKeyCursor

type Direction = internal.Direction

type IndexOptions struct {
	KeyPath    string
	Unique     bool
	MultiEntry bool
}

type Indexer = internal.Indexer

type Index struct {
	def *internal.Index

	Store ReadStore

	h leveldb.Reader

	// Key *func(record interface{}) []byte
	// unique bool
	// multi  bool
}

func (p *Index) Name() string {
	return p.def.Name
}

// are these 3 necessary?
func (p *Index) KeyPath() string {
	return p.def.KeyPath
}

func (p *Index) MultiEntry() bool {
	return p.def.MultiEntry
}

func (p *Index) Unique() bool {
	return p.def.Unique
}

func (p *Index) GetExact(key Key, v interface{}) error {
	primaryKey, err := p.def.GetExact(p.h, key)
	if err != nil {
		return err
	}
	return p.Store.GetExact(primaryKey, v)
}

func (p *Index) Get(query Range, v interface{}) error {
	primaryKey, err := p.def.Get(p.h, query)
	if err != nil {
		return err
	}

	return p.Store.GetExact(primaryKey, v)
}

func (p *Index) GetAll(query Range, limit int, v interface{}) error {
	refs, err := p.def.GetAll(p.h, query, limit)
	if err != nil {
		return err
	}

	return p.Store.GetMulti(refs, v)
}

func (p *Index) GetExactKey(key Key) (Key, error) {
	return p.def.GetExact(p.h, key)
}

func (p *Index) GetKey(query Range) (Key, error) {
	return p.def.Get(p.h, query)
}

func (p *Index) GetAllKeys(query Range, limit int) ([]Key, error) {
	return p.def.GetAll(p.h, query, limit)
}

func (p *Index) Count(query Range) (uint, error) {
	return p.def.Count(p.h, query)
}

func (p *Index) OpenCursor(query Range, dir Direction) (MultiKeyCursor, error) {
	return p.def.GetCursor(p.h, query, dir)
}

func (p *Index) OpenKeyCursor(query Range, dir Direction) (KeyCursor, error) {
	return p.def.GetCursor(p.h, query, dir)
}
