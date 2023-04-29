// (c) 2022-2022, CBOR Schema Group. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"bytes"
	"errors"
	"fmt"

	cbor "github.com/fxamacker/cbor/v2"
)

var (
	_ cbor.Marshaler   = RawKey("")
	_ cbor.Marshaler   = (*RawKey)(nil)
	_ cbor.Unmarshaler = (*RawKey)(nil)
)

type Path []RawKey

func PathFrom(keys ...any) (Path, error) {
	p := make(Path, len(keys))
	for i, key := range keys {
		data, err := cborMarshal(key)
		if err != nil {
			return nil, err
		}

		rk := RawKey(data)
		if err = rk.Valid(); err != nil {
			return nil, err
		}
		p[i] = rk
	}
	return p, nil
}

func PathMustFrom(keys ...any) Path {
	p, err := PathFrom(keys...)
	if err != nil {
		panic(err)
	}
	return p
}

// String returns the Path as CBOR diagnostic notation.
func (p Path) String() string {
	if p == nil {
		return "null"
	}

	buf := &bytes.Buffer{}
	buf.WriteByte('[')
	for i, k := range p {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(k.String())
	}
	buf.WriteByte(']')
	return buf.String()
}

func (p Path) WithInt(i int64) Path {
	return p.WithKey(RawKey(MustMarshal(i)))
}

func (p Path) WithString(s string) Path {
	return p.WithKey(RawKey(MustMarshal(s)))
}

func (p Path) WithKey(key RawKey) Path {
	np := make(Path, len(p)+1)
	copy(np, p)
	np[len(np)-1] = key
	return np
}

func (p Path) Valid() error {
	if p == nil {
		return errors.New("Path: nil pointer")
	}
	for i, k := range p {
		if err := k.Valid(); err != nil {
			return fmt.Errorf("Path: invalid RawKey at %d, %w", i, err)
		}
	}

	return nil
}

// RawKey is a raw encoded CBOR value for map key.
type RawKey string

var minus = RawKey([]byte{0x61, 0x2d}) // "-"

func (k RawKey) isMinus() bool {
	return k == minus
}

func (k RawKey) isIndex() bool {
	if k.isMinus() {
		return true
	}

	switch k.MajorType() {
	default:
		return false

	case MajorUnsignedInt, MajorNegativeInt:
		return true
	}
}

func (k RawKey) MajorType() MajorType {
	return GetMajorType([]byte(k))
}

func (k RawKey) Valid() error {
	switch t := k.MajorType(); t {
	default:
		return fmt.Errorf("RawKey: %q can not be used as map key", t.String())

	case MajorUnsignedInt, MajorNegativeInt, MajorByteString, MajorTextString:
		return cborValid([]byte(k))
	}
}

func (k RawKey) Bytes() []byte {
	return []byte(k)
}

func (k RawKey) Equal(other RawKey) bool {
	return k == other
}

func (k RawKey) Is(other any) bool {
	if data, err := cborMarshal(other); err == nil {
		return k.Equal(RawKey(data))
	}
	return false
}

// String returns the rawKey as CBOR diagnostic notation.
func (k RawKey) String() string {
	return Diagify([]byte(k))
}

func (k RawKey) ToInt() (int64, error) {
	if k.isMinus() {
		return -1, nil
	}

	var i int64
	if err := cborUnmarshal([]byte(k), &i); err != nil {
		return -1, err
	}

	return i, nil
}

// Key returns a string notation as JSON Object key.
func (k RawKey) AsKey() string {
	if k.MajorType() == MajorTextString {
		var val string
		if err := cborUnmarshal([]byte(k), &val); err == nil {
			return val
		}
	}

	return k.String()
}

// MarshalCBOR returns k or CBOR "" if k is nil.
func (k RawKey) MarshalCBOR() ([]byte, error) {
	if len(k) == 0 {
		return []byte{0x60}, nil
	}
	return []byte(k), nil
}

// UnmarshalCBOR creates a copy of data and saves to *k.
func (k *RawKey) UnmarshalCBOR(data []byte) error {
	if k == nil {
		return errors.New("RawKey: nil pointer")
	}

	*k = RawKey(data)
	return k.Valid()
}
