// (c) 2022-2022, CBOR Schema Group. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"fmt"
	"strconv"

	cbor "github.com/fxamacker/cbor/v2"
)

// MajorType is CBOR value's major type.
type MajorType uint8

// Predefined major types.
const (
	MajorUnsignedInt MajorType = 0
	MajorNegativeInt MajorType = 1
	MajorByteString  MajorType = 2
	MajorTextString  MajorType = 3
	MajorArray       MajorType = 4
	MajorMap         MajorType = 5
	MajorTag         MajorType = 6
	MajorPrimitives  MajorType = 7
)

// String returns a string representation of CBORType.
func (t MajorType) String() string {
	switch t {
	case MajorUnsignedInt:
		return "unsigned integer"
	case MajorNegativeInt:
		return "negative integer"
	case MajorByteString:
		return "byte string"
	case MajorTextString:
		return "text string"
	case MajorArray:
		return "array"
	case MajorMap:
		return "map"
	case MajorTag:
		return "tag"
	case MajorPrimitives:
		return "primitives"
	default:
		return "invalid major type " + strconv.Itoa(int(t))
	}
}

// var (
// 	rawCBORNull  = []byte{0xf6}
// 	rawCBORArray = []byte{0x80}
// 	rawCBORMap   = []byte{0xa0}
// )

var (
	decMode, _ = cbor.DecOptions{
		DupMapKey:   cbor.DupMapKeyEnforcedAPF,
		IndefLength: cbor.IndefLengthForbidden,
	}.DecMode()

	encMode, _ = cbor.EncOptions{
		Sort:        cbor.SortBytewiseLexical,
		IndefLength: cbor.IndefLengthForbidden,
	}.EncMode()

	cborUnmarshal = decMode.Unmarshal
	cborValid     = decMode.Valid
	cborMarshal   = encMode.Marshal
)

// SetCBOR set the underlying global CBOR Marshal and Unmarshal functions.
//
//	func init() {
//		var EncMode, _ = cbor.CanonicalEncOptions().EncMode()
//		var DecMode, _ = cbor.DecOptions{
//			DupMapKey:   cbor.DupMapKeyQuiet,
//			IndefLength: cbor.IndefLengthForbidden,
//		}.DecMode()
//
//		cborpatch.SetCBOR(EncMode.Marshal, DecMode.Unmarshal)
//	}
func SetCBOR(
	marshal func(v any) ([]byte, error),
	unmarshal func(data []byte, v any) error,
) {
	cborMarshal = marshal
	cborUnmarshal = unmarshal
}

// RawMessage is a raw encoded CBOR value.
type RawMessage = cbor.RawMessage

// GetMajorType returns the type of a raw encoded CBOR value.
func GetMajorType(data []byte) MajorType {
	switch {
	case len(data) == 0:
		return MajorPrimitives
	default:
		return MajorType((data[0] & 0xe0) >> 5)
	}
}

func MustMarshal(val any) []byte {
	data, err := cborMarshal(val)
	if err != nil {
		panic(err)
	}
	return data
}

// Diagify returns the doc as CBOR diagnostic notation.
// If the doc is a invalid CBOR bytes, it returns the doc with base16 encoding like a byte string.
func Diagify(doc []byte) string {
	if data, err := cbor.Diagnose(doc); err == nil {
		return data
	}

	return fmt.Sprintf("h'%x'", doc)
}
