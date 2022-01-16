package bytewise

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"time"
)

const (
	// VOID byte = 0xf0 // ð
	NULL byte = 0x10 // DLE

	FALSE byte = 0x20 // space
	TRUE  byte = 0x21 // !

	MIN byte = 0x40 // @
	NEG byte = 0x41 // A
	POS byte = 0x42 // B
	MAX byte = 0x43 // C

	DATE_NEG byte = 0x51 // Q
	DATE_POS byte = 0x52 // R

	// BINARY byte = 0x60 / `

	STRING byte = 0x70 // p

	ARRAY byte = 0xa0 // non-breaking space

	END byte = 0x00
)

func Encode(src interface{}) ([]byte, error) {
	res := bytes.NewBuffer([]byte{})
	err := encode(res, src)
	return res.Bytes(), err
}

func encode(buf io.Writer, src interface{}) error {
	t := reflect.TypeOf(src)
	val := reflect.ValueOf(src)
	if src == nil {
		return binary.Write(buf, binary.BigEndian, NULL)
	}
	switch t.Kind() {
	case reflect.Bool:
		if val.Bool() {
			return binary.Write(buf, binary.BigEndian, TRUE)
		}
		return binary.Write(buf, binary.BigEndian, FALSE)
	case reflect.Float64:
		float := src.(float64)
		if float < 0 {
			if float == -math.MaxFloat64 {
				return binary.Write(buf, binary.BigEndian, MIN)
			}
			err := binary.Write(buf, binary.BigEndian, NEG)
			if err != nil {
				return err
			}
			return binary.Write(buf, binary.BigEndian, ^math.Float64bits(-float))
		}
		if float == math.MaxFloat64 {
			return binary.Write(buf, binary.BigEndian, MAX)
		}
		err := binary.Write(buf, binary.BigEndian, POS)
		if err != nil {
			return err
		}
		return binary.Write(buf, binary.BigEndian, math.Float64bits(float))
	case reflect.Struct:
		switch val.Interface().(type) {
		case time.Time:
			t := src.(time.Time).UTC().Unix()
			n := src.(time.Time).Nanosecond()
			if src.(time.Time).Location() != time.UTC {
				return fmt.Errorf("dates should be in UTC")
			}
			marker := DATE_POS
			if t < 0 {
				marker = DATE_NEG
				t = ^(-t)
				n = ^n
			}
			err := binary.Write(buf, binary.BigEndian, marker)
			if err != nil {
				return err
			}
			err = binary.Write(buf, binary.BigEndian, t)
			if err != nil {
				return err
			}
			return binary.Write(buf, binary.BigEndian, int64(n))
		}
		return fmt.Errorf("unknown struct type")
	case reflect.String:
		if strings.ContainsRune(src.(string), 0x00) {
			return fmt.Errorf("strings must not contain null character")
		}
		err := binary.Write(buf, binary.BigEndian, STRING)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.BigEndian, []byte(src.(string)))
		if err != nil {
			return err
		}
		// mark end of string
		return binary.Write(buf, binary.BigEndian, END)
	case reflect.Slice, reflect.Array:
		// e := t.Elem()
		err := binary.Write(buf, binary.BigEndian, ARRAY)
		if err != nil {
			return err
		}
		for i := 0; i < val.Len(); i++ {
			sub := val.Index(i)
			err := encode(buf, sub.Interface())
			if err != nil {
				return fmt.Errorf("problem adding array value %w", err)
			}
			err = binary.Write(buf, binary.BigEndian, END)
			if err != nil {
				return fmt.Errorf("problem adding array deliminator %w", err)
			}
		}
		err = binary.Write(buf, binary.BigEndian, END)
		if err != nil {
			return fmt.Errorf("problem closing array %w", err)
		}
		return nil
	}
	return fmt.Errorf("type %s not supported for encoding", t.Name())
}

func Decode(src []byte) (interface{}, error) {
	buf := bytes.NewBuffer(src)
	var marker byte
	err := binary.Read(buf, binary.BigEndian, &marker)
	if err != nil {
		return nil, err
	}
	return decode(marker, buf)
}

func decode(marker byte, buf io.Reader) (interface{}, error) {
	switch marker {
	case NULL:
		return nil, nil
	case FALSE:
		return false, nil
	case TRUE:
		return true, nil
	case MIN:
		return -math.MaxFloat64, nil
	case NEG:
		var neg uint64
		err := binary.Read(buf, binary.BigEndian, &neg)
		if err != nil {
			return nil, err
		}
		return -math.Float64frombits(^neg), nil
	case POS:
		var pos uint64
		err := binary.Read(buf, binary.BigEndian, &pos)
		if err != nil {
			return nil, err
		}
		return math.Float64frombits(pos), nil
	case MAX:
		return math.MaxFloat64, nil
	case DATE_NEG, DATE_POS:
		var sec int64
		err := binary.Read(buf, binary.BigEndian, &sec)
		if err != nil {
			return nil, err
		}
		var nano int64
		err = binary.Read(buf, binary.BigEndian, &nano)
		if err != nil {
			return nil, err
		}
		if marker == DATE_NEG {
			sec = -(^sec)
			nano = ^nano
		}
		return time.Unix(sec, nano).UTC(), nil
	// case BINARY:
	case STRING:
		// loop by byte until we get to an END
		out := []byte{}
		var char byte
		err := binary.Read(buf, binary.BigEndian, &char)
		if err != nil {
			return nil, fmt.Errorf("problem initializing string reader %w", err)
		}
		for {
			if char == END {
				break
			}
			out = append(out, char)
			err = binary.Read(buf, binary.BigEndian, &char)
			if err != nil {
				return nil, fmt.Errorf("problem processing string %w", err)
			}
		}
		return string(out), nil
	case ARRAY:
		var out = make([]interface{}, 0)
		var next byte
		err := binary.Read(buf, binary.BigEndian, &next)
		if err != nil {
			return out, fmt.Errorf("problem initializing array %w", err)
		}
		for {
			if next == END {
				break
			}
			val, err := decode(next, buf)
			if err != nil {
				return out, fmt.Errorf("problem parsing array entry %w", err)
			}
			out = append(out, val)
			// remove the deliminator
			err = binary.Read(buf, binary.BigEndian, &next)
			if err != nil {
				return out, fmt.Errorf("problem removing deliminator %w", err)
			}
			if next != END {
				return out, fmt.Errorf("invalid array")
			}
			// take one step forward
			err = binary.Read(buf, binary.BigEndian, &next)
			if err != nil {
				return out, fmt.Errorf("problem initializing next array item %w", err)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("unrecognized token %s", string(marker))
}
