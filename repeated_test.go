package pbvector

import (
	"testing"

	"github.com/koykov/vector"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestRepeated(t *testing.T) {
	t.Run("varint", func(t *testing.T) {
		var b []byte
		for _, v := range []uint64{1, 2, 3} {
			b = protowire.AppendTag(b, 4, protowire.VarintType)
			b = protowire.AppendVarint(b, v)
		}
		vec := mustParse(t, b)
		arr := vec.Root().Get("4")
		if arr.Type() != vector.TypeArray {
			t.Fatalf("type = %v, want array", arr.Type())
		}
		if arr.Limit() != 3 {
			t.Fatalf("len = %d, want 3", arr.Limit())
		}
		for i, want := range []uint64{1, 2, 3} {
			got, err := vec.GetUint("4", string(rune('0'+i)))
			if err != nil || got != want {
				t.Fatalf("element %d = %d,%v want %d", i, got, err, want)
			}
		}
	})

	t.Run("interleaved with message", func(t *testing.T) {
		var msg []byte
		msg = protowire.AppendTag(msg, 1, protowire.VarintType)
		msg = protowire.AppendVarint(msg, 9)

		var b []byte
		b = protowire.AppendTag(b, 4, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
		b = protowire.AppendTag(b, 5, protowire.BytesType)
		b = protowire.AppendBytes(b, msg)
		b = protowire.AppendTag(b, 4, protowire.VarintType)
		b = protowire.AppendVarint(b, 2)

		vec := mustParse(t, b)
		arr := vec.Root().Get("4")
		if arr.Type() != vector.TypeArray || arr.Limit() != 2 {
			t.Fatalf("field 4 arr = %v limit %d", arr.Type(), arr.Limit())
		}
		if v, err := vec.GetUint("4", "0"); err != nil || v != 1 {
			t.Fatalf("arr[0] = %d,%v", v, err)
		}
		if v, err := vec.GetUint("4", "1"); err != nil || v != 2 {
			t.Fatalf("arr[1] = %d,%v", v, err)
		}
		if m := vec.Root().Get("5"); m.Type() != vector.TypeObject {
			t.Fatalf("field 5 = %v, want object", m.Type())
		}
	})
}
