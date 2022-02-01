package internal

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Database struct {
	*leveldb.DB
	Name    string            `json:"name"`
	Version uint              `json:"version"`
	Stores  map[string]*Store `json:"-"`
}

func (p *Database) StoreNames() []string {
	keys := make([]string, 0, len(p.Stores))
	for k := range p.Stores {
		keys = append(keys, k)
	}
	return keys
}

func (p *Database) UpdateDefinition(r *leveldb.Transaction) error {
	def, _ := json.Marshal(p)
	return r.Put(Key{}.forCore(), def, nil)
}

func (p *Database) CreateStore(r *leveldb.Transaction, spec Store) (*Store, error) {
	val, _ := json.Marshal(spec)

	key := Key{"store", spec.Name}.forCore()

	err := r.Put(key, val, nil)
	if err != nil {
		return nil, err
	}

	store := NewStore(p, spec)

	p.Stores[store.Name] = store
	return store, nil
}

func (p *Database) CreateIndex(r *leveldb.Transaction, spec Index) (*Index, error) {
	val, _ := json.Marshal(spec)

	key := Key{"index", spec.Name}.forCore()

	err := r.Put(key, val, nil)
	if err != nil {
		return nil, err
	}

	index := NewIndex(p, spec)

	return index, nil
}

func (p *Database) DeleteIndex(r *leveldb.Transaction, idx *Index) error {
	err := idx.Clear(r)
	if err != nil {
		return err
	}
	return r.Delete(Key{"idx", idx.Name}.forCore(), nil)
}

func (p *Database) GetExact(r leveldb.Reader, k []byte) ([]byte, error) {
	return r.Get(k, nil)
}

func (p *Database) Get(r leveldb.Reader, k util.Range) ([]byte, []byte, error) {
	iter := r.NewIterator(&k, nil)
	defer iter.Release()

	if !iter.First() {
		return nil, nil, fmt.Errorf("record not found")
	}
	return iter.Key(), iter.Value(), nil
}

func (p *Database) GetIter(r leveldb.Reader, k util.Range, cb func(key []byte, val []byte) bool) {
	iter := r.NewIterator(&k, nil)
	for iter.Next() {
		if !cb(iter.Key(), iter.Value()) {
			break
		}
	}
	iter.Release()
}

func (p *Database) GetMulti(r leveldb.Reader, keys [][]byte, cb func(row []byte) error) error {
	for _, k := range keys {
		val, err := p.GetExact(r, k)
		if err != nil {
			return err
		}
		err = cb(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Database) Count(r leveldb.Reader, k util.Range) (uint, error) {
	var count uint = 0
	iter := r.NewIterator(&k, nil)
	defer iter.Release()

	for iter.Next() {
		count += 1
	}

	return count, nil
}

// hydrate loads the known store and index definitions into the database instance
func (p *Database) Hydrate() error {
	p.Stores = make(map[string]*Store)
	iter := p.NewIterator(Range{Start: &Key{"store"}, Prefix: true}.forCore(), nil)
	for iter.Next() {
		var spec Store
		err := json.Unmarshal(iter.Value(), &spec)
		if err != nil {
			return err
		}
		store := NewStore(p, spec)

		p.Stores[store.Name] = store
	}
	iter.Release()

	iter = p.NewIterator(Range{Start: &Key{"index"}, Prefix: true}.forCore(), nil)
	for iter.Next() {
		var spec Index
		err := json.Unmarshal(iter.Value(), &spec)
		if err != nil {
			return err
		}

		index := NewIndex(p, spec)

		if _, ok := p.Stores[index.StoreName]; ok {
			p.Stores[index.StoreName].Indexes[index.Name] = index
		}
	}
	iter.Release()
	return nil
}

func OpenDatabase(name string, path string) (*Database, error) {

	h, err := leveldb.OpenFile(filepath.Join(path, name), &opt.Options{})
	if err != nil {
		return nil, err
	}

	def := Database{h, name, 0, make(map[string]*Store)}

	data, err := h.Get(Key{}.forCore(), nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			return nil, err
		}
	} else {
		err = json.Unmarshal(data, &def)
		if err != nil {
			return nil, err
		}
	}
	return &def, nil
}
