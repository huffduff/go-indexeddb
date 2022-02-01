package internal

import (
	"fmt"

	"github.com/huffduff/go-indexeddb/bytewise"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Key []interface{}

func (p Key) forCore() []byte {
	v := append([]interface{}{"core"}, p...)
	return bytewise.MustEncode(v...)
}

func rawKey(src []byte) ([]interface{}, error) {
	data, err := bytewise.Decode(src)
	if err != nil {
		return nil, err
	}
	raw, ok := data.([]interface{})
	if !ok {
		err = fmt.Errorf("not a valid key")
	}
	return raw, err
}

func fromCore(src []byte) (Key, error) {
	data, err := bytewise.Decode(src)
	if err != nil {
		return nil, err
	}
	coerced := data.([]interface{})
	if coerced[0] != "core" {
		return nil, fmt.Errorf("key is not a valid core key")
	}
	var p Key = coerced[1:]
	return p, nil
}

func (p Key) forIndex(i *Index, id interface{}) ([]byte, error) {
	key := append([]interface{}{"idx", i.Name}, p...)
	if !i.Unique {
		key = append(key, id)
	}
	return bytewise.Encode(key)
}

func fromIndex(i *Index, src []byte) (string, Key, error) {
	data, err := bytewise.Decode(src)
	if err != nil {
		return "", nil, err
	}
	coerced := data.([]interface{})
	if coerced[0] != "idx" {
		return "", nil, fmt.Errorf("key is not a valid index key")
	}
	name, ok := coerced[1].(string)
	if !ok {
		return name, nil, fmt.Errorf("key does not contain a valid index name")
	}
	last := len(coerced)
	if i.Unique {
		last = -1
	}
	var p Key = coerced[2:last]
	return name, p, nil
}

func (p Key) forStore(s *Store) ([]byte, error) {
	key := append([]interface{}{"data", s.Name}, p...)
	return bytewise.Encode(key)
}

func fromStore(src []byte) (string, Key, error) {
	data, err := bytewise.Decode(src)
	if err != nil {
		return "", nil, err
	}
	coerced := data.([]interface{})
	if coerced[0] != "data" {
		return "", nil, fmt.Errorf("key is not a valid index key")
	}
	name, ok := coerced[1].(string)
	if !ok {
		return name, nil, fmt.Errorf("key does not contain a valid store name")
	}
	var p Key = coerced[2:]
	return name, p, nil
}

// Next generates the key that would be the minimal next key
func (p Key) Next() Key {
	out := make(Key, len(p)+1)
	copy(out, p)
	return out
}

// Stop generates the key that would mark the end of this key as a prefix
func (p Key) Stop() Key {
	out := make(Key, len(p))
	copy(out, p)
done:
	for i := len(out) - 1; i >= 0; i-- {
		asBytes := bytewise.MustEncode(out[i])
		for j := len(asBytes) - 1; j >= 0; j-- {
			if asBytes[j] < byte(0xff) {
				asBytes[j] = asBytes[j] + 1
				out[i], _ = bytewise.Decode(asBytes)
				break done
			}
		}
	}
	return out
}

type Range struct {
	Start          *Key
	Limit          *Key
	StartExclusive bool
	LimitInclusive bool
	Prefix         bool
}

func (p Range) prep() Range {
	var start = Key{}
	if p.Start != nil {
		start = *p.Start
	}
	if p.StartExclusive {
		start = start.Next()
	}

	var limit = Key{}
	if p.Prefix {
		limit = start.Stop()
	} else if p.Limit != nil {
		limit := *p.Limit
		if p.LimitInclusive {
			limit = limit.Next()
		}
	}
	return Range{Start: &start, Limit: &limit}
}

func (p Range) forStore(s *Store) (util.Range, error) {
	out := util.Range{}
	var err error

	fixed := p.prep()

	out.Start, err = fixed.Start.forStore(s)
	if err != nil {
		return out, err
	}

	out.Limit, err = fixed.Limit.forStore(s)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (p Range) forIndex(i *Index) (util.Range, error) {
	out := util.Range{}
	var err error

	fixed := p.prep()

	out.Start, err = fixed.Start.forIndex(i, nil)
	if err != nil {
		return out, err
	}

	out.Limit, err = fixed.Limit.forIndex(i, nil)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (p Range) forCore() *util.Range {
	fixed := p.prep()

	return &util.Range{
		Start: fixed.Start.forCore(),
		Limit: fixed.Limit.forCore(),
	}
}
