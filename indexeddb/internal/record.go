package internal

import "encoding/json"

type Record struct {
	IndexKeys map[string][][]byte `json:"indexKeys"`
	Value     json.RawMessage     `json:"value"`
}
