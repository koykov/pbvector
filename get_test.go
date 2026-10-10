package pbvector

import (
	"math"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
)

func TestGet(t *testing.T) {
	t.Run("int uint varint", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 300)
		vec := mustParse(t, b)
		if v, err := vec.GetInt("1"); err != nil || v != 300 {
			t.Fatalf("GetInt = %d,%v", v, err)
		}
		if v, err := vec.GetUint("1"); err != nil || v != 300 {
			t.Fatalf("GetUint = %d,%v", v, err)
		}
	})

	t.Run("uint above max int64", func(t *testing.T) {
		big := uint64(math.MaxInt64) + 1
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, big)
		vec := mustParse(t, b)
		if v, err := vec.GetUint("1"); err != nil || v != big {
			t.Fatalf("GetUint = %d,%v want %d", v, err, big)
		}
		if v, err := vec.GetInt("1"); err != nil || v != int64(big) {
			t.Fatalf("GetInt = %d,%v want wrap %d", v, err, int64(big))
		}
	})

	t.Run("fixed and float", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.Fixed32Type)
		b = protowire.AppendFixed32(b, math.Float32bits(1.5))
		b = protowire.AppendTag(b, 2, protowire.Fixed64Type)
		b = protowire.AppendFixed64(b, math.Float64bits(2.5))
		vec := mustParse(t, b)
		if v, err := vec.GetFloat("1"); err != nil || v != 1.5 {
			t.Fatalf("GetFloat fixed32 = %v,%v", v, err)
		}
		if v, err := vec.GetFloat("2"); err != nil || v != 2.5 {
			t.Fatalf("GetFloat fixed64 = %v,%v", v, err)
		}
	})

	t.Run("int fixed32 negative", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.Fixed32Type)
		b = protowire.AppendFixed32(b, 0xFFFFFFFF)
		vec := mustParse(t, b)
		if v, err := vec.GetInt("1"); err != nil || v != -1 {
			t.Fatalf("GetInt = %d,%v want -1", v, err)
		}
	})

	t.Run("bool", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
		b = protowire.AppendTag(b, 2, protowire.VarintType)
		b = protowire.AppendVarint(b, 0)
		vec := mustParse(t, b)
		if !vec.GetBool("1") || vec.GetBool("2") {
			t.Fatalf("GetBool = %v, %v", vec.GetBool("1"), vec.GetBool("2"))
		}
	})

	t.Run("dot and not found", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 7)
		vec := mustParse(t, b)
		if v, err := vec.DotInt("1"); err != nil || v != 7 {
			t.Fatalf("DotInt = %d,%v", v, err)
		}
		if _, err := vec.GetInt("9"); err == nil {
			t.Fatal("expected error for missing field")
		}
		if WireType(vec.Get("1")) != wireVarint {
			t.Fatalf("WireType = %d", WireType(vec.Get("1")))
		}
	})
}
