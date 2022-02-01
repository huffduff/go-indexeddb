package internal

import (
	"bytes"
	"testing"

	"github.com/huffduff/go-indexeddb/bytewise"
)

var ordered []Key = []Key{
	{"core"},
	{"core", "index", "bar"},
	{"core", "index", "foo"},
	{"core", "store", "bar"},
	{"core", "store", "foo"},
	{"idx", "bar", "a"},
	{"idx", "bar", "b"},
	{"idx", "foo", 3.0, "a"},
	{"idx", "foo", 3.0, "b"},
	{"data", "bar", "record"},
	{"data", "bar", "record 1"},
	{"data", "bar", "record 2"},
}

var orderedIndex [][]byte = make([][]byte, len(ordered))

func init() {
	for i, k := range ordered {
		orderedIndex[i] = bytewise.MustEncode(k...)
	}
}

func TestForCore(t *testing.T) {
	if !bytes.Equal(Key{}.forCore(), bytewise.MustEncode("core")) {
		t.Error("keys do not match")
	}
	if !bytes.Equal(Key{"store"}.forCore(), bytewise.MustEncode("core", "store")) {
		t.Error("store keys do not match")
	}
	if !bytes.Equal(Key{"index"}.forCore(), bytewise.MustEncode("core", "index")) {
		t.Error("index keys do not match")
	}
}

func TestForStore(t *testing.T) {
	store := &Store{Name: "foo"}
	k, err := Key{"row 1"}.forStore(store)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(k, bytewise.MustEncode("data", "foo", "row 1")) {
		t.Error("keys do not match")
	}
}

func TestForUniqueIndex(t *testing.T) {
	idx := &Index{Name: "foo", Unique: true}
	k, err := Key{3.0}.forIndex(idx, nil)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(k, bytewise.MustEncode("idx", "foo", 3.0)) {
		t.Error("keys do not match")
	}

	k, err = Key{3.0}.forIndex(idx, "some id")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(k, bytewise.MustEncode("idx", "foo", 3.0)) {
		t.Error("keys do not match")
	}
}

func TestForNonUniqueIndex(t *testing.T) {
	idx := &Index{Name: "foo", Unique: false}
	k1, err := Key{3.0}.forIndex(idx, nil)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(k1, bytewise.MustEncode("idx", "foo", 3.0, nil)) {
		t.Error("keys do not match")
	}

	k2, err := Key{3.0}.forIndex(idx, "some id")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(k2, bytewise.MustEncode("idx", "foo", 3.0, "some id")) {
		t.Error("keys do not match")
	}

	if bytes.Equal(k1, k2) {
		t.Error("unique keys shouldn't match")
	}
}

func TestNext(t *testing.T) {
	k := Key{"a"}
	n := k.Next()
	b := Key{"a", 0.0}

	kBytes := bytewise.MustEncode(k...)
	nBytes := bytewise.MustEncode(n...)
	bBytes := bytewise.MustEncode(b...)

	if bytes.Compare(kBytes, nBytes) != -1 {
		t.Error("next is not of higher order")
	}
	if bytes.Compare(nBytes, bBytes) != -1 {
		t.Error("next should be before any prefixed versions")
	}
}

func TestStop(t *testing.T) {
	k := Key{"aaa"}
	n := Key{"aaa", 0.0}
	l := k.Stop()
	b := Key{"aab"}

	kBytes := bytewise.MustEncode(k...)
	nBytes := bytewise.MustEncode(n...)
	lBytes := bytewise.MustEncode(l...)
	bBytes := bytewise.MustEncode(b...)

	if match := bytes.Compare(kBytes, lBytes); match != -1 {
		t.Errorf("stop (%v) should be after the source (%v) %d", kBytes, lBytes, match)
	}
	if match := bytes.Compare(nBytes, lBytes); match != -1 {
		t.Errorf("stop (%v) should be after any source derivative (%v) %d", nBytes, lBytes, match)
	}
	if match := bytes.Compare(lBytes, bBytes); match <= -1 {
		t.Errorf("stop (%v) should be before or at the next natural order (%v) %d", lBytes, bBytes, match)
	}

}
