package pbvector

import (
	"testing"

	"github.com/koykov/vector"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestCrossCheck(t *testing.T) {
	t.Run("traversal", func(t *testing.T) {
		var b []byte
		b = protowire.AppendTag(b, 1, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
		b = protowire.AppendTag(b, 2, protowire.BytesType)
		b = protowire.AppendString(b, "abc")
		b = protowire.AppendTag(b, 5, protowire.Fixed32Type)
		b = protowire.AppendFixed32(b, 3)
		b = protowire.AppendTag(b, 5, protowire.Fixed32Type)
		b = protowire.AppendFixed32(b, 4)

		vec := mustParse(t, b)
		seen := map[protowire.Number]bool{}
		for len(b) > 0 {
			num, _, n := protowire.ConsumeField(b)
			if n < 0 {
				t.Fatalf("protowire: %v", protowire.ParseError(n))
			}
			seen[num] = true
			b = b[n:]
		}
		for num := range seen {
			key := itoa(int(num))
			if vec.Root().Get(key).Type() == vector.TypeNull {
				t.Fatalf("field %d missing in tree", num)
			}
		}
	})
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
