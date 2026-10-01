package pbvector

import (
	"encoding/binary"
	"math"

	"github.com/koykov/vector"
)

func WireType(n *vector.Node) int {
	if n == nil {
		return -1
	}
	v := n.Value()
	switch {
	case v.CheckBit(pbWireVarint):
		return wireVarint
	case v.CheckBit(pbWireFixed64):
		return wireFixed64
	case v.CheckBit(pbWireLen):
		return wireLen
	case v.CheckBit(pbWireFixed32):
		return wireFixed32
	}
	return -1
}

func notFound(n *vector.Node) bool {
	return n == nil || n.Type() == vector.TypeNull || n.Type() == vector.TypeUnknown
}

func (vec *Vector) decodeInt(n *vector.Node) (int64, error) {
	if notFound(n) {
		return 0, vector.ErrNotFound
	}
	if n.Type() != vector.TypeNumber {
		return 0, vector.ErrIncompatType
	}
	raw := n.Value().RawBytes()
	switch WireType(n) {
	case wireVarint:
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return int64(v), nil
	case wireFixed64:
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return int64(binary.LittleEndian.Uint64(raw)), nil
	case wireFixed32:
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return int64(int32(binary.LittleEndian.Uint32(raw))), nil
	}
	return 0, ErrIncompatibleWireType
}

func (vec *Vector) decodeUint(n *vector.Node) (uint64, error) {
	if notFound(n) {
		return 0, vector.ErrNotFound
	}
	if n.Type() != vector.TypeNumber {
		return 0, vector.ErrIncompatType
	}
	raw := n.Value().RawBytes()
	switch WireType(n) {
	case wireVarint:
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return v, nil
	case wireFixed64:
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return binary.LittleEndian.Uint64(raw), nil
	case wireFixed32:
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return uint64(binary.LittleEndian.Uint32(raw)), nil
	}
	return 0, ErrIncompatibleWireType
}

func (vec *Vector) decodeFloat(n *vector.Node) (float64, error) {
	if notFound(n) {
		return 0, vector.ErrNotFound
	}
	if n.Type() != vector.TypeNumber {
		return 0, vector.ErrIncompatType
	}
	raw := n.Value().RawBytes()
	switch WireType(n) {
	case wireVarint:
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return float64(v), nil
	case wireFixed64:
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(raw)), nil
	case wireFixed32:
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(raw))), nil
	}
	return 0, ErrIncompatibleWireType
}

func (vec *Vector) decodeBool(n *vector.Node) bool {
	if notFound(n) || n.Type() != vector.TypeNumber {
		return false
	}
	if WireType(n) != wireVarint {
		return false
	}
	v, _, ok := decodeVarint(n.Value().RawBytes())
	return ok && v != 0
}

func (vec *Vector) GetInt(keys ...string) (int64, error) {
	return vec.decodeInt(vec.Vector.Get(keys...))
}

func (vec *Vector) GetUint(keys ...string) (uint64, error) {
	return vec.decodeUint(vec.Vector.Get(keys...))
}

func (vec *Vector) GetFloat(keys ...string) (float64, error) {
	return vec.decodeFloat(vec.Vector.Get(keys...))
}

func (vec *Vector) GetBool(keys ...string) bool {
	return vec.decodeBool(vec.Vector.Get(keys...))
}

func (vec *Vector) GetIntPS(path, separator string) (int64, error) {
	return vec.decodeInt(vec.Vector.GetPS(path, separator))
}

func (vec *Vector) GetUintPS(path, separator string) (uint64, error) {
	return vec.decodeUint(vec.Vector.GetPS(path, separator))
}

func (vec *Vector) GetFloatPS(path, separator string) (float64, error) {
	return vec.decodeFloat(vec.Vector.GetPS(path, separator))
}

func (vec *Vector) GetBoolPS(path, separator string) bool {
	return vec.decodeBool(vec.Vector.GetPS(path, separator))
}

func (vec *Vector) DotInt(path string) (int64, error) {
	return vec.decodeInt(vec.Vector.Dot(path))
}

func (vec *Vector) DotUint(path string) (uint64, error) {
	return vec.decodeUint(vec.Vector.Dot(path))
}

func (vec *Vector) DotFloat(path string) (float64, error) {
	return vec.decodeFloat(vec.Vector.Dot(path))
}

func (vec *Vector) DotBool(path string) bool {
	return vec.decodeBool(vec.Vector.Dot(path))
}
