package pbvector

import (
	"testing"

	"github.com/koykov/vector"
)

func TestVector(t *testing.T) {
	t.Run("new vector", func(t *testing.T) {
		vec := NewVector()
		if !vec.CheckBit(vector.FlagInit) {
			t.Fatal("expected FlagInit to be set")
		}
	})

	t.Run("acquire release", func(t *testing.T) {
		vec := Acquire()
		if !vec.CheckBit(vector.FlagInit) {
			t.Fatal("expected FlagInit on pooled vector")
		}
		Release(vec)
		vec2 := Acquire()
		if vec2 == nil || !vec2.CheckBit(vector.FlagInit) {
			t.Fatal("expected reused pooled vector to be initialized")
		}
		Release(vec2)
	})

	t.Run("reset reuse", func(t *testing.T) {
		src := []byte{0x08, 0x2A}
		vec := NewVector()
		if err := vec.ParseCopy(src); err != nil {
			t.Fatalf("first parse: %v", err)
		}
		vec.Reset()
		if err := vec.ParseCopy(src); err != nil {
			t.Fatalf("second parse after Reset: %v", err)
		}
		if v, err := vec.GetInt("1"); err != nil || v != 42 {
			t.Fatalf("after Reset field 1 = %d,%v want 42", v, err)
		}
	})
}
