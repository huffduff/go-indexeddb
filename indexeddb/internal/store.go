package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Store struct {
	*Database

	Name          string            `json:"name"`
	KeyPath       string            `json:"keyPath,omitempty"`
	AutoIncrement bool              `json:"autoIncrement"`
	Indexes       map[string]*Index `json:"-"`
}

func (p *Store) IndexNames() []string {
	keys := make([]string, 0, len(p.Indexes))
	for k := range p.Indexes {
		keys = append(keys, k)
	}
	return keys
}

func (p *Store) put(tr *leveldb.Transaction, primaryKey []byte, existingIdx map[string][][]byte, value interface{}) error {
	var err error

	b := &leveldb.Batch{}

	record := Record{}

	// idxKeys := make([]Key, 0)
	// get a reflected version once
	// ref := reflect.ValueOf(value)
	for idxName := range p.Indexes {
		idx := p.Indexes[idxName]

		var erase [][]byte
		copy(erase, existingIdx[idxName])

		keys := idx.Keys(primaryKey, value)

		record.IndexKeys[idxName] = make([][]byte, len(keys))

		for i, key := range keys {
			k, _ := key.forIndex(idx, key)
			b.Put(k, primaryKey)
			record.IndexKeys[idxName][i] = k
			// scrub from the erase list
			i := 0
			for _, h := range erase {
				if bytes.Equal(h, k) {
					erase[i] = h
					i++
					break
				}
			}
			erase = erase[:i]
		}
		for _, e := range erase {
			b.Delete(e)
		}
	}
	record.Value, err = json.Marshal(value)
	if err != nil {
		return err
	}
	val, _ := json.Marshal(record)

	b.Put(primaryKey, val)

	return tr.Write(b, nil)
}

func (p *Store) Put(tr *leveldb.Transaction, key Key, value interface{}) error {
	primaryKey, err := key.forStore(p)
	if err != nil {
		return err
	}

	data, err := tr.Get(primaryKey, nil)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return err
	}

	var record Record
	json.Unmarshal(data, &record)

	return p.put(tr, primaryKey, record.IndexKeys, value)
}

func (p *Store) Add(tr *leveldb.Transaction, key Key, value interface{}) error {
	primaryKey, err := key.forStore(p)
	if err != nil {
		return err
	}

	exists, err := tr.Has(primaryKey, nil)
	if !errors.Is(err, leveldb.ErrNotFound) {
		return err
	}

	if exists {
		return fmt.Errorf("record already exists")
	}

	return p.put(tr, primaryKey, nil, value)
}

func (p *Store) Delete(tr leveldb.Transaction, key Key) error {
	primaryKey, err := key.forStore(p)
	if err != nil {
		return err
	}

	ref, err := tr.Get(primaryKey, nil)
	if err != nil {
		return err
	}

	var record Record
	json.Unmarshal(ref, &record)

	b := &leveldb.Batch{}

	for _, keys := range record.IndexKeys {
		for _, key := range keys {
			b.Delete(key)
		}
	}
	b.Delete(primaryKey)

	return tr.Write(b, nil)
}

func (p *Store) Clear(r leveldb.Transaction) error {
	q, _ := Range{Start: &Key{}, Prefix: true}.forStore(p)
	b := &leveldb.Batch{}
	iter := r.NewIterator(&q, nil)
	for iter.Next() {
		b.Delete(iter.Key())
	}
	iter.Release()
	return p.Database.Write(b, nil)
}

func (p *Store) GetExact(r leveldb.Reader, key Key, v interface{}) error {
	primaryKey, err := key.forStore(p)
	if err != nil {
		return err
	}

	data, err := p.Database.GetExact(r, primaryKey)
	if err != nil {
		return err
	}

	var record Record
	err = json.Unmarshal(data, &record)
	if err != nil {
		return err
	}

	return json.Unmarshal(record.Value, v)
}

func (p *Store) Get(r leveldb.Reader, query Range, v interface{}) error {
	q, err := query.forStore(p)
	if err != nil {
		return err
	}

	_, data, err := p.Database.Get(r, q)
	if err != nil {
		return err
	}

	var record Record
	err = json.Unmarshal(data, &record)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (p *Store) GetAll(r leveldb.Reader, query Range, limit int, v interface{}) error {

	q, err := query.forStore(p)
	if err != nil {
		return err
	}

	// FIXME this is terribly inefficient
	sizes, _ := p.Database.SizeOf([]util.Range{q})

	out := bytes.NewBuffer(make([]byte, 0, sizes[0]))

	i := 0

	out.WriteByte('[')
	p.Database.GetIter(r, q, func(key, val []byte) bool {
		if i > 0 {
			out.WriteByte(',')
		}
		out.Write(val)
		i++
		return i != limit
	})
	out.WriteByte(']')

	return json.NewDecoder(out).Decode(v)
}

// FIXME this is terribly inefficient
func (p *Store) GetMulti(r leveldb.Reader, keys []Key, v interface{}) error {
	out := bytes.NewBuffer(make([]byte, 0, len(keys)*1024))

	i := 0

	out.WriteByte('[')
	for _, key := range keys {
		if i > 0 {
			out.WriteByte(',')
		}
		primaryKey, _ := key.forStore(p)
		val, _ := p.Database.GetExact(r, primaryKey)
		out.Write(val)
		i++
	}
	out.WriteByte(']')

	return json.NewDecoder(out).Decode(v)
}

func (p *Store) GetKey(r leveldb.Reader, query Range) (Key, error) {
	q, err := query.forStore(p)
	if err != nil {
		return nil, err
	}

	key, _, err := p.Database.Get(r, q)
	if err != nil {
		return nil, err
	}

	_, k, err := fromStore(key)
	return k, err
}

func (p *Store) GetAllKeys(r leveldb.Reader, query Range, limit int) ([]Key, error) {
	q, err := query.forStore(p)
	if err != nil {
		return nil, err
	}

	keys := make([]Key, 0)
	p.Database.GetIter(r, q, func(key, _ []byte) bool {
		_, val, e := fromStore(key)
		if e != nil {
			err = e
			return false
		}
		keys = append(keys, val)
		return true
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (p *Store) Count(r leveldb.Reader, query Range) (uint, error) {
	q, err := query.forStore(p)
	if err != nil {
		return 0, err
	}
	return p.Database.Count(r, q)
}

func (p *Store) GetCursor(r leveldb.Reader, query Range, dir Direction) (*StoreCursor, error) {
	q, err := query.forStore(p)
	if err != nil {
		return nil, err
	}
	iter := r.NewIterator(&q, nil)
	return &StoreCursor{p, BaseCursor{iter: iter, direction: dir}}, nil
}

func NewStore(h *Database, spec Store) *Store {
	return &Store{h, spec.Name, spec.KeyPath, spec.AutoIncrement, make(map[string]*Index)}
}
