package bytewise

import (
	"bytes"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

func Simple(t *testing.T, val interface{}) {
	buf, err := Encode(val)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v", string(buf))
	res, err := Decode(buf)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(res, val) {
		t.Errorf("%v != %v", res, val)
	}
}

func TestNull(t *testing.T) {
	Simple(t, nil)
}

func TestFalse(t *testing.T) {
	Simple(t, false)
}

func TestTrue(t *testing.T) {
	Simple(t, true)
}

func TestMin(t *testing.T) {
	Simple(t, -math.MaxFloat64)
}

func TestMax(t *testing.T) {
	Simple(t, math.MaxFloat64)
}

func TestNumbers(t *testing.T) {
	Simple(t, -math.SmallestNonzeroFloat64)
	Simple(t, 0.0)
	Simple(t, math.SmallestNonzeroFloat64)
}

func TestDates(t *testing.T) {
	Simple(t, time.Date(1960, time.January, 1, 12, 30, 59, 499, time.UTC))
	Simple(t, time.Now().UTC())
}

func TestStrings(t *testing.T) {
	Simple(t, "foo")
}

func TestArrays(t *testing.T) {
	Simple(t, []interface{}{true, nil, 8.8, "bar"})
}

func TestSorts(t *testing.T) {
	sorted := []interface{}{
		nil,
		-4.0,
		-0.304958230,
		0.0,
		0.304958230,
		4.0,
		"bar",
		"baz",
		"foo",
		[]interface{}{0.0, 0.0, "foo"},
		[]interface{}{0.0, 1.0, "foo"},
		[]interface{}{0.0, 1.0, "foo", 0.0},
		[]interface{}{0.0, 1.0, "foo", 1.0},
		[]interface{}{0.0, "bar", "baz"},
		[]interface{}{0.0, "foo"},
		[]interface{}{0.0, "foo", "bar"},
		[]interface{}{0.0, "foo", []interface{}{}},
		[]interface{}{0.0, "foo", []interface{}{"bar"}},
		[]interface{}{0.0, "foo", []interface{}{"bar"}, []interface{}{}},
		[]interface{}{0.0, "foo", []interface{}{"bar"}, []interface{}{"foo"}},
		[]interface{}{0.0, "foo", []interface{}{"bar", "baz"}},
		[]interface{}{1.0, "bar", "baz"},
		[]interface{}{1.0, "bar", "baz"},
		[]interface{}{"foo", "bar", "baz"},
		[]interface{}{"foo", []interface{}{"bar", "baz"}},
		[]interface{}{"foo", []interface{}{"bar", []interface{}{"baz"}}},
	}

	mixed := make([]interface{}, len(sorted))
	copy(mixed, sorted)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(mixed), func(i, j int) { mixed[i], mixed[j] = mixed[j], mixed[i] })

	keys := make([][]byte, len(mixed))
	for i, v := range mixed {
		val, err := Encode(v)
		if err != nil {
			t.Errorf("problem encoding row %d: %v", i, err)
		}
		keys[i] = val
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return bytes.Compare(keys[i], keys[j]) == -1
	})

	for i, v := range keys {
		d, err := Decode(v)
		if err != nil {
			t.Errorf("problem decoding row %d: %v", i, err)
		}
		if !reflect.DeepEqual(sorted[i], d) {
			t.Errorf("%d: %v != %v", i, sorted[i], d)
		}
	}
}
