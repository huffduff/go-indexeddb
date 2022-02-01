package indexeddb

import (
	"github.com/huffduff/go-indexeddb/indexeddb/internal"
)

type Key = internal.Key

type Range = internal.Range

// Lowerbound generates a range starting at the selected key
func LowerBound(lower Key, open bool) Range {
	return Range{
		Start:          &lower,
		StartExclusive: open,
	}
}

// UpperBound generates a range starting from the beginning and ending at a select key
func UpperBound(upper Key, open bool) Range {
	return Range{
		Limit:          &upper,
		LimitInclusive: !open,
	}
}

// Only Generates a range with optional inclusivity flags
func Bound(lower, upper Key, lowerOpen, upperOpen bool) Range {
	return Range{
		Start:          &lower,
		StartExclusive: lowerOpen,
		Limit:          &upper,
		LimitInclusive: !upperOpen,
	}
}

func All() Range {
	return Range{}
}

// Only generates a range with an exact match
func Only(key internal.Key) Range {
	next := key.Next()
	return Range{
		Start: &key,
		Limit: &next,
	}
}

// Prefix generates a range satisfying the Key as a prefix
func Prefix(key Key) Range {
	return Range{
		Start:  &key,
		Prefix: true,
	}
}
