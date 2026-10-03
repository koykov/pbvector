package pbvector

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/koykov/vector"
)

type Codec struct{}

func (Codec) Decode(p *vector.Byteptr) ([]byte, error) { return p.RawBytes(), nil }

func (Codec) DecodeInt(p *vector.Byteptr) (int64, error) {
	raw := p.RawBytes()
	switch {
	case p.CheckBit(pbWireVarint):
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return int64(v), nil
	case p.CheckBit(pbWireFixed64):
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return int64(binary.LittleEndian.Uint64(raw)), nil
	case p.CheckBit(pbWireFixed32):
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return int64(int32(binary.LittleEndian.Uint32(raw))), nil
	}
	return 0, ErrIncompatibleWireType
}

func (Codec) DecodeUint(p *vector.Byteptr) (uint64, error) {
	raw := p.RawBytes()
	switch {
	case p.CheckBit(pbWireVarint):
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return v, nil
	case p.CheckBit(pbWireFixed64):
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return binary.LittleEndian.Uint64(raw), nil
	case p.CheckBit(pbWireFixed32):
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return uint64(binary.LittleEndian.Uint32(raw)), nil
	}
	return 0, ErrIncompatibleWireType
}

func (Codec) DecodeFloat(p *vector.Byteptr) (float64, error) {
	raw := p.RawBytes()
	switch {
	case p.CheckBit(pbWireVarint):
		v, _, ok := decodeVarint(raw)
		if !ok {
			return 0, ErrVarintOverflow
		}
		return float64(v), nil
	case p.CheckBit(pbWireFixed64):
		if len(raw) < 8 {
			return 0, ErrTruncated
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(raw)), nil
	case p.CheckBit(pbWireFixed32):
		if len(raw) < 4 {
			return 0, ErrTruncated
		}
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(raw))), nil
	}
	return 0, ErrIncompatibleWireType
}

func (Codec) DecodeBool(p *vector.Byteptr) (bool, error) {
	raw := p.RawBytes()
	switch {
	case p.CheckBit(pbWireVarint):
		v, _, ok := decodeVarint(raw)
		if !ok {
			return false, ErrVarintOverflow
		}
		return v != 0, nil
	}
	return false, ErrIncompatibleWireType
}

func (Codec) Beautify(_ io.Writer, _ *vector.Node) error {
	return vector.ErrNotImplement
}

func (Codec) Marshal(_ io.Writer, _ *vector.Node) error {
	return vector.ErrNotImplement
}
