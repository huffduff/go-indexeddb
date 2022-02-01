package internal

import (
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
)

type Indexer interface {
	Keys(indexName string) []Key
}

type Index struct {
	*Database

	Name       string `json:"name"`
	StoreName  string `json:"store"`
	KeyPath    string `json:"keypath,omitempty"`
	Unique     bool   `json:"unique"`
	MultiEntry bool   `json:"multiEntry"`
}

// TODO This seems pricey
func (p *Index) Keys(id interface{}, record interface{}) []Key {
	// if our record implements its own Keys method, use it
	if m, ok := record.(Indexer); ok {
		return p.enforceUnique(id, m.Keys(p.Name))
	}

	r := reflect.ValueOf(record)
	val := r.FieldByName(p.KeyPath)
	if p.MultiEntry {
		s := val.Len()
		keys := make([]Key, s)
		for i := 0; i < s; i++ {
			keys[i] = Key{val.Index(s).Interface()}
		}
		return p.enforceUnique(id, keys)
	}

	return p.enforceUnique(id, []Key{{val.Interface()}})
}

func (p *Index) enforceUnique(id interface{}, keys []Key) []Key {
	if !p.Unique {
		for i := range keys {
			keys[i] = append(keys[i], id)
		}
	}
	return keys
}

func (p *Index) GetExact(r leveldb.Reader, key Key) (Key, error) {
	indexKey, err := key.forIndex(p, nil)
	if err != nil {
		return nil, err
	}

	val, err := p.Database.GetExact(r, indexKey)
	if err != nil {
		return nil, err
	}

	_, primaryKey, err := fromStore(val)
	return primaryKey, err
}

func (p *Index) Get(r leveldb.Reader, query Range) (Key, error) {
	q, err := query.forIndex(p)
	if err != nil {
		return nil, err
	}

	_, val, err := p.Database.Get(r, q)
	if err != nil {
		return nil, err
	}

	_, primaryKey, err := fromStore(val)
	return primaryKey, err
}

func (p *Index) GetAll(r leveldb.Reader, query Range, limit int) ([]Key, error) {

	q, err := query.forIndex(p)
	if err != nil {
		return nil, err
	}

	out := make([]Key, limit)

	i := 0

	p.Database.GetIter(r, q, func(key, val []byte) bool {
		// TODO handle corrupt key error?
		_, primaryKey, _ := fromStore(val)
		out = append(out, primaryKey)
		i++
		return i != limit
	})

	return out, nil
}

func (p *Index) Count(r leveldb.Reader, query Range) (uint, error) {
	q, err := query.forIndex(p)
	if err != nil {
		return 0, err
	}
	return p.Database.Count(r, q)
}

func (p *Index) GetCursor(r leveldb.Reader, query Range, dir Direction) (*IndexCursor, error) {
	q, err := query.forIndex(p)
	if err != nil {
		return nil, err
	}
	iter := r.NewIterator(&q, nil)
	return &IndexCursor{p, BaseCursor{iter: iter, direction: dir}}, nil
}

func (p *Index) Clear(r *leveldb.Transaction) error {
	q, _ := Range{Start: &Key{}, Prefix: true}.forIndex(p)
	b := &leveldb.Batch{}
	iter := r.NewIterator(&q, nil)
	for iter.Next() {
		b.Delete(iter.Key())
	}
	iter.Release()
	return p.Database.Write(b, nil)
}

func NewIndex(h *Database, spec Index) *Index {
	return &Index{h, spec.Name, spec.StoreName, spec.KeyPath, spec.Unique, spec.MultiEntry}
}
