package pbvector

import (
	"errors"
	"runtime"
	"testing"

	"github.com/koykov/vector"
	"google.golang.org/protobuf/encoding/protowire"
)

func mustParse(t *testing.T, src []byte) *Vector {
	t.Helper()
	vec := NewVector()
	if err := vec.ParseCopy(src); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return vec
}

func TestParser(t *testing.T) {
	t.Run("scalars", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 42)
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, "hello")
		b = protowire.AppendTag(b, 3, protowire.Fixed64Type)
		b = protowire.AppendFixed64(b, 7)

		vec := mustParse(t, b)
		root := vec.Root()
		if root.Type() != vector.TypeObject {
			t.Fatalf("root type = %v, want object", root.Type())
		}
		if root.Limit() != 3 {
			t.Fatalf("root children = %d, want 3", root.Limit())
		}
		n1 := root.Get("1")
		if n1.Type() != vector.TypeNumber || n1.KeyString() != "1" {
			t.Fatalf("field 1 = %+v (key %q)", n1.Type(), n1.KeyString())
		}
		if string(root.Get("2").RawBytes()) != "hello" {
			t.Fatalf("field 2 raw = %q", root.Get("2").RawBytes())
		}
	})

	t.Run("empty", func(t *testing.T) {
		vec := NewVector()
		if err := vec.Parse(nil); err == nil {
			t.Fatal("expected error for empty source")
		}
	})

	t.Run("truncated", func(t *testing.T) {
		vec := NewVector()
		if err := vec.ParseCopy([]byte{0x08, 0x80}); err == nil {
			t.Fatal("expected error for truncated source")
		}
	})

	t.Run("recursion limit", func(t *testing.T) {
		msg := []byte{0x08, 0x01}
		for i := 0; i < 4; i++ {
			var b []byte
			b = protowire.AppendTag(b, 1, protowire.BytesType)
			b = protowire.AppendBytes(b, msg)
			msg = b
		}
		vec := NewVector()
		vec.maxDepth = 3
		if err := vec.ParseCopy(msg); !errors.Is(err, ErrRecursionLimit) {
			t.Fatalf("ParseCopy = %v, want ErrRecursionLimit", err)
		}
	})

	t.Run("key stability across gc", func(t *testing.T) {
		const n = 5000
		var b []byte
		for i := 1; i <= n; i++ {
			b = protowire.AppendTag(b, protowire.Number(i), protowire.VarintType)
			b = protowire.AppendVarint(b, uint64(i))
		}
		vec := mustParse(t, b)
		runtime.GC()
		runtime.GC()
		if got := vec.Root().Get("5000"); got.Type() == vector.TypeNull {
			t.Fatal("field 5000 missing after GC")
		}
		if got := vec.Root().Get("1"); got.KeyString() != "1" {
			t.Fatalf("field 1 key = %q, want %q", got.KeyString(), "1")
		}
		if got := vec.Root().Limit(); got != n {
			t.Fatalf("root children = %d, want %d", got, n)
		}
	})
}
