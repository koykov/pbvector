package pbvector

import (
	"bytes"
	"errors"
	"testing"

	"github.com/koykov/vector"
)

func TestWire(t *testing.T) {
	t.Run("decode varint", func(t *testing.T) {
		cases := []struct {
			name string
			in   []byte
			val  uint64
			n    int
			ok   bool
		}{
			{"one", []byte{0x08}, 8, 1, true},
			{"multi", []byte{0xAC, 0x02}, 300, 2, true},
			{"max10", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}, 0xFFFFFFFFFFFFFFFF, 10, true},
			{"truncated", []byte{0x80}, 0, 0, false},
			{"empty", nil, 0, 0, false},
			{"overflow11", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 0, 0, false},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				v, n, ok := decodeVarint(c.in)
				if ok != c.ok || (ok && (v != c.val || n != c.n)) {
					t.Fatalf("decodeVarint(%v) = (%d,%d,%v), want (%d,%d,%v)", c.in, v, n, ok, c.val, c.n, c.ok)
				}
			})
		}
	})

	t.Run("scan message", func(t *testing.T) {
		src := []byte{
			0x08, 0xAC, 0x02,
			0x11, 1, 0, 0, 0, 0, 0, 0, 0,
			0x1A, 0x02, 'h', 'i',
		}
		fields, err := scanMessage(src, 0, len(src))
		if err != nil {
			t.Fatalf("expected valid message, got %v", err)
		}
		if len(fields) != 3 {
			t.Fatalf("got %d fields, want 3", len(fields))
		}
		if fields[0].num != 1 || fields[0].wire != wireVarint {
			t.Fatalf("field0 = %+v", fields[0])
		}
		if fields[1].num != 2 || fields[1].wire != wireFixed64 || fields[1].length != 8 {
			t.Fatalf("field1 = %+v", fields[1])
		}
		if fields[2].num != 3 || fields[2].wire != wireLen || fields[2].length != 2 || fields[2].off != 14 {
			t.Fatalf("field2 = %+v", fields[2])
		}

		if _, err := scanMessage([]byte{0x08, 0x80}, 0, 2); !errors.Is(err, ErrTruncated) {
			t.Fatalf("truncated message: got %v, want ErrTruncated", err)
		}
		if _, err := scanMessage([]byte{0x1A, 0x05, 'h', 'i'}, 0, 4); !errors.Is(err, ErrTruncated) {
			t.Fatalf("length beyond end: got %v, want ErrTruncated", err)
		}
		if _, err := scanMessage([]byte{0x00}, 0, 1); !errors.Is(err, ErrUnexpectedWireType) {
			t.Fatalf("field number 0: got %v, want ErrUnexpectedWireType", err)
		}
	})

	t.Run("error check", func(t *testing.T) {
		cases := []struct {
			name string
			src  []byte
			want error
		}{
			{"unexpected-wire", []byte{0x0B}, ErrUnexpectedWireType},
			{"varint-overflow", []byte{0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0D}, ErrVarintOverflow},
			{"truncated", []byte{0x08, 0x80}, ErrTruncated},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				vec := NewVector()
				err := vec.ParseCopy(c.src)
				if !errors.Is(err, c.want) {
					t.Fatalf("ParseCopy(% X) = %v, want %v", c.src, err, c.want)
				}
			})
		}
	})

	t.Run("overflow", func(t *testing.T) {
		t.Run("len", func(t *testing.T) {
			lenb := append(bytes.Repeat([]byte{0xFF}, 9), 0x01)
			b := append([]byte{0x1A}, append(lenb, 0xAA)...)
			vec := NewVector()
			if err := vec.ParseCopy(b); err == nil {
				t.Fatal("expected error for overflowing length")
			}
			vec.Root().Each(func(idx int, node *vector.Node) {
				if len(node.RawBytes()) > 5 {
					t.Fatalf("oversized node value: %d bytes", len(node.RawBytes()))
				}
			})
		})

		t.Run("varint", func(t *testing.T) {
			b := []byte{0x08, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0D}
			vec := NewVector()
			if err := vec.ParseCopy(b); err == nil {
				t.Fatal("expected error for overlong varint")
			}
		})
	})
}
