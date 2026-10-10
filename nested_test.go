package pbvector

import (
	"testing"

	"github.com/koykov/vector"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestNested(t *testing.T) {
	t.Run("nested", func(t *testing.T) {
		var inner []byte
		inner = protowire.AppendTag(inner, 1, protowire.VarintType)
		inner = protowire.AppendVarint(inner, 5)
		inner = protowire.AppendTag(inner, 2, protowire.BytesType)
		inner = protowire.AppendString(inner, "x")

		var outer []byte
		outer = protowire.AppendTag(outer, 1, protowire.BytesType)
		outer = protowire.AppendBytes(outer, inner)

		vec := mustParse(t, outer)
		obj := vec.Root().Get("1")
		if obj.Type() != vector.TypeObject {
			t.Fatalf("nested field type = %v, want object", obj.Type())
		}
		if obj.Get("2").Type() != vector.TypeString || string(obj.Get("2").RawBytes()) != "x" {
			t.Fatalf("nested field 2 = %q", obj.Get("2").RawBytes())
		}
	})

	t.Run("empty len is empty object", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendBytes(b, nil)

		vec := mustParse(t, b)
		n := vec.Root().Get("1")
		if n.Type() != vector.TypeObject {
			t.Fatalf("empty payload type = %v, want object", n.Type())
		}
		if n.Limit() != 0 {
			t.Fatalf("empty object children = %d, want 0", n.Limit())
		}
	})

	t.Run("non-message bytes is leaf", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.BytesType)
		b = protowire.AppendBytes(b, []byte{0xFF, 0xFE})

		vec := mustParse(t, b)
		if vec.Root().Get("1").Type() != vector.TypeString {
			t.Fatalf("expected leaf string, got %v", vec.Root().Get("1").Type())
		}
	})

	t.Run("empty object no phantom child", func(t *testing.T) {
		src := []byte{0x0A, 0x00, 0x12, 0x02, 0x08, 0x09}
		vec := NewVector()
		if err := vec.ParseCopy(src); err != nil {
			t.Fatalf("parse: %v", err)
		}
		one := vec.Root().Get("1")
		if one.Type() != vector.TypeObject {
			t.Fatalf("field 1 type = %v, want object", one.Type())
		}
		if n := len(one.ChildrenIndices()); n != 0 {
			t.Fatalf("field 1 children = %d, want 0", n)
		}
		if got := one.Get("1"); got.Type() != vector.TypeNull {
			t.Fatalf("field 1.1 type = %v, want null", got.Type())
		}
		if v, err := vec.GetInt("2", "1"); err != nil || v != 9 {
			t.Fatalf("field 2.1 = %d,%v want 9", v, err)
		}
	})

	t.Run("nested first parse no panic", func(t *testing.T) {
		src := []byte{0x0A, 0x00, 0x12, 0x02, 0x08, 0x09}
		vec := NewVector()
		if err := vec.ParseCopy(src); err != nil {
			t.Fatalf("parse: %v", err)
		}
		_ = vec.Root().Get("2").Get("1")
	})
}
