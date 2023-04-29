// (c) 2022-2022, CBOR Schema Group. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/x448/float16"
)

var (
	_ json.Marshaler   = (*Tag)(nil)
	_ cbor.Marshaler   = (*Tag)(nil)
	_ cbor.Unmarshaler = (*Tag)(nil)

	_ json.Marshaler   = (*Node)(nil)
	_ cbor.Marshaler   = (*Node)(nil)
	_ cbor.Unmarshaler = (*Node)(nil)
)

// SimpleType represents the CBOR Schema simple types.
type SimpleType uint8

const (
	InvalidType SimpleType = iota
	NullType
	BoolType
	UintType
	IntType
	Float16Type
	Float32Type
	Float64Type
	BytesType
	TextType
	ArrayType
	MapType
	TagType
)

// Node represents a lazy parsing CBOR document.
type Node struct {
	mt  MajorType
	st  SimpleType
	raw RawMessage
	val any
}

type Array = []*Node

type Map = map[RawKey]*Node

type Tag struct {
	num  uint64
	node *Node
}

// 65535: always invalid tag
// https://www.ietf.org/archive/id/draft-bormann-cbor-notable-tags-02.html#name-implementation-aids
const InvalidTag = 65535

func (t *Tag) IsValid() bool {
	switch t.num {
	default:
		return t.node != nil
	case InvalidTag, 4294967295, 18446744073709551615:
		return false
	}
}

func (t *Tag) Number() uint64 {
	return t.num
}

func (t *Tag) Node() *Node {
	return t.node
}

func (t *Tag) MarshalCBOR() ([]byte, error) {
	if t == nil || t.node == nil {
		return []byte{0xf6}, nil
	}

	return cbor.RawTag{Number: t.num, Content: t.node.raw}.MarshalCBOR()
}

func (t *Tag) UnmarshalCBOR(data []byte) error {
	if t == nil {
		return errors.New("Tag: UnmarshalCBOR on nil pointer")
	}

	val := &cbor.RawTag{}
	if err := val.UnmarshalCBOR(data); err != nil {
		return fmt.Errorf("Tag: UnmarshalCBOR error, %w", err)
	}

	t.num = val.Number
	t.node = newNode(val.Content)
	return nil
}

func (t *Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{t.num, t.node})
}

// NewNode returns a new Node with the given raw encoded CBOR document.
// A nil or empty raw document is equal to CBOR null.
func NewNode(doc RawMessage) (*Node, error) {
	if err := cborValid(doc); err != nil {
		return nil, err
	}

	return newNode(doc), nil
}

func newNode(doc RawMessage) *Node {
	n := &Node{raw: doc}
	n.initType()
	return n
}

func (n Node) MajorType() MajorType {
	return n.mt
}

func (n Node) SimpleType() SimpleType {
	return n.st
}

func (n Node) IsNull() bool {
	return n.st == NullType
}

func (n Node) IsBool() bool {
	return n.st == BoolType
}

func (n Node) IsUint() bool {
	return n.st == UintType
}

func (n Node) IsInt() bool {
	return n.st == UintType || n.st == IntType
}

func (n Node) IsFloat16() bool {
	return n.st == Float16Type
}

func (n Node) IsFloat32() bool {
	return n.st == Float32Type
}

func (n Node) IsFloat64() bool {
	return n.st == Float64Type
}

func (n Node) IsFloat() bool {
	return n.IsFloat64() || n.IsFloat32() || n.IsFloat16()
}

func (n Node) IsBytes() bool {
	return n.st == BytesType
}

func (n Node) IsText() bool {
	return n.st == TextType
}

func (n Node) IsArray() bool {
	return n.st == ArrayType
}

func (n Node) IsMap() bool {
	return n.st == MapType
}

func (n Node) IsTag() bool {
	return n.st == TagType
}

func (n Node) As(val any) error {
	return cborUnmarshal(n.raw, val)
}

func (n *Node) AsBool() bool {
	val, ok := n.val.(bool)
	return ok && val
}

func (n *Node) AsUint() uint64 {
	var val uint64
	if n.st == UintType {
		if v, ok := n.val.(uint64); ok {
			return v
		}
		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsInt() int64 {
	var val int64
	if n.st == UintType {
		return int64(n.AsUint())
	} else if n.st == IntType {
		if v, ok := n.val.(int64); ok {
			return v
		}
		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsFloat16() float16.Float16 {
	var val float16.Float16
	if n.IsFloat16() {
		if v, ok := n.val.(float16.Float16); ok {
			return v
		}

		var v16 uint16
		_ = n.As(&v16)
		val = float16.Frombits(v16)
		n.val = val
	}
	return val
}

func (n *Node) AsFloat32() float32 {
	var val float32
	if n.IsFloat16() {
		return n.AsFloat16().Float32()
	} else if n.IsFloat32() {
		if v, ok := n.val.(float32); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsFloat64() float64 {
	var val float64
	if n.IsFloat16() || n.IsFloat32() {
		return float64(n.AsFloat32())
	} else if n.IsFloat64() {
		if v, ok := n.val.(float64); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}

	return val
}

func (n *Node) AsBytes() []byte {
	var val []byte
	if n.IsBytes() {
		if v, ok := n.val.([]byte); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsText() string {
	var val string
	if n.IsText() {
		if v, ok := n.val.(string); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsArray() Array {
	var val Array
	if n.IsArray() {
		if v, ok := n.val.(Array); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsMap() Map {
	var val Map
	if n.IsMap() {
		if v, ok := n.val.(Map); ok {
			return v
		}

		_ = n.As(&val)
		n.val = val
	}
	return val
}

func (n *Node) AsTag() Tag {
	if n.st == TagType {
		val := &cbor.RawTag{}
		if n.As(val) == nil {
			return Tag{val.Number, newNode(val.Content)}
		}
	}

	return Tag{InvalidTag, n}
}

func (n Node) Raw() RawMessage {
	return n.raw
}

func (n *Node) Lookup(p Path) (*Node, error) {
	if len(p) == 0 {
		return n, nil
	}

	k := p[0]
	if err := k.Valid(); err != nil {
		return nil, err
	}

	switch n.st {
	case ArrayType:
		if k.isIndex() {
			idx, err := k.ToInt()
			if err != nil {
				return nil, err
			}
			arr := n.AsArray()
			if idx < 0 {
				idx += int64(len(arr))
			}
			if idx < 0 || idx >= int64(len(arr)) {
				return nil, errors.New("Node: index out of range")
			}
			return arr[idx].Lookup(p[1:])
		}
	case MapType:
		obj := n.AsMap()
		if v, ok := obj[k]; ok {
			return v.Lookup(p[1:])
		}
	}

	return nil, errors.New("Node: not found")
}

func (n *Node) Valid() error {
	if n == nil || n.raw == nil {
		return errors.New("Node: Valid on nil pointer")
	}

	if err := cborValid(n.raw); err != nil {
		return err
	}

	if n.st == InvalidType {
		return errors.New("Node: invalid SimpleType")
	}

	return nil
}

func (n *Node) MarshalCBOR() ([]byte, error) {
	if n == nil || n.raw == nil {
		return []byte{0xf6}, nil
	}

	return n.raw, nil
}

func (n *Node) UnmarshalCBOR(data []byte) error {
	if n == nil {
		return errors.New("Node: UnmarshalCBOR on nil pointer")
	}

	n.raw = append((n.raw)[0:0], data...)
	n.initType()
	return nil
}

func (n *Node) MarshalJSON() ([]byte, error) {
	switch n.st {
	case NullType:
		return []byte("null"), nil

	case BoolType:
		return json.Marshal(n.AsBool())

	case UintType:
		return json.Marshal(n.AsUint())

	case IntType:
		return json.Marshal(n.AsInt())

	case Float16Type:
		return json.Marshal(n.AsFloat16().Float32())

	case Float32Type:
		return json.Marshal(n.AsFloat32())

	case Float64Type:
		return json.Marshal(n.AsFloat64())

	case BytesType:
		return json.Marshal(base64.RawURLEncoding.EncodeToString(n.AsBytes()))

	case TextType:
		return json.Marshal(n.AsText())

	case ArrayType:
		return json.Marshal(n.AsArray())

	case MapType:
		cmap := n.AsMap()
		if cmap == nil {
			return []byte("null"), nil
		}

		jmap := make(map[string]*Node, len(cmap))
		for k, v := range cmap {
			jmap[k.AsKey()] = v
		}
		return json.Marshal(jmap)

	case TagType:
		if tag := n.AsTag(); tag.IsValid() {
			return tag.MarshalJSON()
		}
	}

	var val any
	if err := cborUnmarshal(n.raw, &val); err != nil {
		return nil, err
	}

	return json.Marshal(val)
}

func (n *Node) initType() {
	if len(n.raw) == 0 {
		n.mt = MajorPrimitives
		n.st = InvalidType
		return
	}

	n.mt = GetMajorType(n.raw)
	switch n.mt {
	case MajorUnsignedInt:
		n.st = UintType
	case MajorNegativeInt:
		n.st = IntType
	case MajorByteString:
		n.st = BytesType
	case MajorTextString:
		n.st = TextType
	case MajorArray:
		n.st = ArrayType
	case MajorMap:
		n.st = MapType
	case MajorTag:
		n.st = TagType
	case MajorPrimitives:
		switch n.raw[0] & 0x1f {
		case 20:
			n.st = BoolType
			n.val = false
		case 21:
			n.st = BoolType
			n.val = true
		case 22, 23:
			n.st = NullType
			n.val = nil
		case 25:
			n.st = Float16Type
		case 26:
			n.st = Float32Type
		case 27:
			n.st = Float64Type
		default:
			n.st = InvalidType
		}
	}
}

// func copyBytes(data []byte) []byte {
// 	if data == nil {
// 		return nil
// 	}
// 	b := make([]byte, len(data))
// 	copy(b, data)
// 	return b
// }
