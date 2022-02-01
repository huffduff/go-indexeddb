package internal

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/iterator"
)

type KeyCursor interface {
	Source()
	Direction() Direction
	Key() (Key, error)
	PrimaryKey() Key
	Request()
	Advance(count int) bool
	Continue() bool
	ContinueTo(key Key) error
}

type Cursor interface {
	KeyCursor
	Value(v interface{}) error
	Delete() error
	Update(val interface{}) error
}

type MultiKeyCursor interface {
	KeyCursor
	ContinuePrimaryKey(key Key, primaryKey Key) error
}

var _ KeyCursor = (*StoreCursor)(nil)
var _ Cursor = (*StoreCursor)(nil)

var _ KeyCursor = (*IndexCursor)(nil)
var _ MultiKeyCursor = (*IndexCursor)(nil)

type Direction string

const (
	NEXT       Direction = "next"
	PREV       Direction = "prev"
	NEXTUNIQUE Direction = "nextunique"
	PREVUNIQUE Direction = "prevunique"
)

type BaseCursor struct {
	iter       iterator.Iterator
	direction  Direction
	primaryKey Key // ??
}

func (p *BaseCursor) Source() {

}

func (p *BaseCursor) Direction() Direction {
	return p.direction
}

func (p *BaseCursor) PrimaryKey() Key {
	return p.primaryKey
}

func (p *BaseCursor) Request() {}

func (p *BaseCursor) Advance(count int) bool {
	for i := 0; i < count; i++ {
		if !p.Continue() {
			return false
		}
	}
	return true
}

func (p *BaseCursor) Continue() bool {
	switch p.direction {
	case PREV:
		return p.iter.Prev()
	case PREVUNIQUE:
		// FIXME
		return p.iter.Prev()
	case NEXTUNIQUE:
		// FIXME
		return p.iter.Next()
	}
	return p.iter.Next()
}

type StoreCursor struct {
	store *Store
	BaseCursor
}

func (p *StoreCursor) Key() (Key, error) {
	val := p.iter.Key()
	_, key, err := fromStore(val)
	return key, err
}

func (p *StoreCursor) ContinueTo(key Key) error {
	k, err := key.forStore(p.store)
	if err != nil {
		return err
	}
	if !p.iter.Seek(k) {
		return fmt.Errorf("key not found")
	}
	return nil
}

func (p *StoreCursor) Value(val interface{}) error {
	return json.Unmarshal(p.iter.Value(), val)
}

func (p *StoreCursor) Delete() error {
	// TODO
	return nil
}

func (p *StoreCursor) Update(val interface{}) error {
	// TODO
	return nil
}

type IndexCursor struct {
	idx *Index
	BaseCursor
}

func (p *IndexCursor) Key() (Key, error) {
	val := p.iter.Value()
	// TODO should this return the index key or the doc key?
	_, key, err := fromStore(val)
	return key, err

}

func (p *IndexCursor) ContinueTo(key Key) error {
	k, err := key.forIndex(p.idx, key)
	if err != nil {
		return err
	}
	if !p.iter.Seek(k) {
		return fmt.Errorf("key not found")
	}
	return nil
}

// IndexCursor only
func (p *IndexCursor) ContinuePrimaryKey(key Key, primaryKey Key) error {
	// TODO needs store access
	return nil
}
