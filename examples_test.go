package pbvector

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/koykov/vector"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func fVarint(dst []byte, num protowire.Number, v uint64) []byte {
	dst = protowire.AppendTag(dst, num, protowire.VarintType)
	return protowire.AppendVarint(dst, v)
}

func fStr(dst []byte, num protowire.Number, s string) []byte {
	dst = protowire.AppendTag(dst, num, protowire.BytesType)
	return protowire.AppendString(dst, s)
}

func fBytes(dst []byte, num protowire.Number, b []byte) []byte {
	dst = protowire.AppendTag(dst, num, protowire.BytesType)
	return protowire.AppendBytes(dst, b)
}

func fMsg(dst []byte, num protowire.Number, msg []byte) []byte {
	dst = protowire.AppendTag(dst, num, protowire.BytesType)
	return protowire.AppendBytes(dst, msg)
}

func mustMarshal(m proto.Message) []byte {
	b, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func mustVecInt(t *testing.T, vec *Vector, keys ...string) int64 {
	t.Helper()
	v, err := vec.GetInt(keys...)
	require.NoError(t, err)
	return v
}

func mustVecUint(t *testing.T, vec *Vector, keys ...string) uint64 {
	t.Helper()
	v, err := vec.GetUint(keys...)
	require.NoError(t, err)
	return v
}

func mustVecFloat(t *testing.T, vec *Vector, keys ...string) float64 {
	t.Helper()
	v, err := vec.GetFloat(keys...)
	require.NoError(t, err)
	return v
}

func mustVecBool(t *testing.T, vec *Vector, keys ...string) bool {
	t.Helper()
	return vec.GetBool(keys...)
}

func mustVecString(t *testing.T, vec *Vector, keys ...string) string {
	t.Helper()
	return vec.GetString(keys...)
}

func mustNodeInt(t *testing.T, n *vector.Node) int64 {
	t.Helper()
	v, err := DecodeInt(n)
	require.NoError(t, err)
	return v
}

func mustNodeFloat(t *testing.T, n *vector.Node) float64 {
	t.Helper()
	v, err := DecodeFloat(n)
	require.NoError(t, err)
	return v
}

func mustNodeBool(n *vector.Node) bool {
	return DecodeBool(n)
}

type stage struct {
	name  string
	build func() []byte
	check func(t *testing.T, vec *Vector)
}

var stages = []stage{
	{
		name: "Timestamp",
		build: func() []byte {
			return mustMarshal(timestamppb.New(time.Unix(1700000000, 123)))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, int64(1700000000), mustVecInt(t, vec, "1"))
			require.Equal(t, int64(123), mustVecInt(t, vec, "2"))
		},
	},
	{
		name: "Duration",
		build: func() []byte {
			return mustMarshal(durationpb.New(90*time.Second + 500*time.Millisecond))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, int64(90), mustVecInt(t, vec, "1"))
			require.Equal(t, int64(500000000), mustVecInt(t, vec, "2"))
		},
	},
	{
		name: "DoubleValue",
		build: func() []byte {
			return mustMarshal(wrapperspb.Double(1.5))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, 1.5, mustVecFloat(t, vec, "1"))
		},
	},
	{
		name: "FloatValue",
		build: func() []byte {
			return mustMarshal(wrapperspb.Float(1.5))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, 1.5, mustVecFloat(t, vec, "1"))
		},
	},
	{
		name: "Int64Value",
		build: func() []byte {
			return mustMarshal(wrapperspb.Int64(42))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, int64(42), mustVecInt(t, vec, "1"))
		},
	},
	{
		name: "UInt64Value",
		build: func() []byte {
			return mustMarshal(wrapperspb.UInt64(math.MaxUint64))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, uint64(math.MaxUint64), mustVecUint(t, vec, "1"))
		},
	},
	{
		name: "BoolValue",
		build: func() []byte {
			return mustMarshal(wrapperspb.Bool(true))
		},
		check: func(t *testing.T, vec *Vector) {
			require.True(t, mustVecBool(t, vec, "1"))
		},
	},
	{
		name: "StringValue",
		build: func() []byte {
			return mustMarshal(wrapperspb.String("hello"))
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, "hello", mustVecString(t, vec, "1"))
			require.Equal(t, vector.TypeString, vec.Root().Get("1").Type())
		},
	},
	{
		name: "BytesValue",
		build: func() []byte {
			return mustMarshal(wrapperspb.Bytes([]byte{0x00, 0x01, 0x02}))
		},
		check: func(t *testing.T, vec *Vector) {
			n := vec.Root().Get("1")
			require.Equal(t, vector.TypeString, n.Type())
			require.Equal(t, []byte{0x00, 0x01, 0x02}, n.RawBytes())
		},
	},
	{
		name: "StringValue/empty",
		build: func() []byte {
			var b []byte
			b = fBytes(b, 1, nil)
			return b
		},
		check: func(t *testing.T, vec *Vector) {
			n := vec.Root().Get("1")
			require.Equal(t, vector.TypeObject, n.Type())
			require.Equal(t, 0, n.Limit())
		},
	},
	{
		name: "Any",
		build: func() []byte {
			msg, err := anypb.New(timestamppb.New(time.Unix(111, 222)))
			if err != nil {
				panic(err)
			}
			return mustMarshal(msg)
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, "type.googleapis.com/google.protobuf.Timestamp", mustVecString(t, vec, "1"))
			require.Equal(t, vector.TypeObject, vec.Root().Get("2").Type())
			require.Equal(t, int64(111), mustVecInt(t, vec, "2", "1"))
			require.Equal(t, int64(222), mustVecInt(t, vec, "2", "2"))
		},
	},
	{
		name: "FieldMask",
		build: func() []byte {
			return mustMarshal(&fieldmaskpb.FieldMask{Paths: []string{"user.display_name", "photo"}})
		},
		check: func(t *testing.T, vec *Vector) {
			arr := vec.Root().Get("1")
			require.Equal(t, vector.TypeArray, arr.Type())
			require.Equal(t, 2, arr.Limit())
			require.Equal(t, "user.display_name", mustVecString(t, vec, "1", "0"))
			require.Equal(t, "photo", mustVecString(t, vec, "1", "1"))
		},
	},
	{
		name: "Struct",
		build: func() []byte {
			s, err := structpb.NewStruct(map[string]any{
				"name": "alice",
				"age":  42.0,
				"ok":   true,
				"none": nil,
				"list": []any{"x", 1.0},
			})
			if err != nil {
				panic(err)
			}
			return mustMarshal(s)
		},
		check: func(t *testing.T, vec *Vector) {
			arr := vec.Root().Get("1")
			require.Equal(t, vector.TypeArray, arr.Type())
			require.Equal(t, 5, arr.Limit())

			byKey := make(map[string]*vector.Node, arr.Limit())
			for i := 0; i < arr.Limit(); i++ {
				e := arr.At(i)
				byKey[string(e.Get("1").RawBytes())] = e
			}
			for _, key := range []string{"name", "age", "ok", "none", "list"} {
				require.NotNil(t, byKey[key], "missing key %q", key)
			}

			name := byKey["name"].Get("2")
			require.Equal(t, vector.TypeString, name.Get("3").Type())
			require.Equal(t, "alice", string(name.Get("3").RawBytes()))

			require.Equal(t, 42.0, mustNodeFloat(t, byKey["age"].Get("2").Get("2")))
			require.True(t, mustNodeBool(byKey["ok"].Get("2").Get("4")))
			require.Equal(t, int64(0), mustNodeInt(t, byKey["none"].Get("2").Get("1")))

			lv := byKey["list"].Get("2").Get("6")
			require.Equal(t, vector.TypeObject, lv.Type())
			ls := lv.Get("1")
			require.Equal(t, vector.TypeArray, ls.Type())
			require.Equal(t, 2, ls.Limit())
			require.Equal(t, "x", string(ls.At(0).Get("3").RawBytes()))
			require.Equal(t, 1.0, mustNodeFloat(t, ls.At(1).Get("2")))
		},
	},
	{
		name: "AddressBook",
		build: func() []byte {
			phoneA := fStr(nil, 1, "+1-555-0100")
			phoneB := fStr(nil, 1, "+1-555-0101")
			phoneB = fVarint(phoneB, 2, 2)

			ts1 := fVarint(nil, 1, 1700000000)
			var p1 []byte
			p1 = fStr(p1, 1, "Alice")
			p1 = fVarint(p1, 2, 1234)
			p1 = fStr(p1, 3, "alice@example.com")
			p1 = fMsg(p1, 4, phoneA)
			p1 = fMsg(p1, 4, phoneB)
			p1 = fMsg(p1, 5, ts1)

			ts2 := fVarint(nil, 1, 1700000123)
			ts2 = fVarint(ts2, 2, 456000000)
			var p2 []byte
			p2 = fStr(p2, 1, "Bob")
			p2 = fVarint(p2, 2, 5678)
			p2 = fStr(p2, 3, "bob@example.com")
			p2 = fMsg(p2, 5, ts2)

			var b []byte
			b = fMsg(b, 1, p1)
			b = fMsg(b, 1, p2)
			return b
		},
		check: func(t *testing.T, vec *Vector) {
			people := vec.Root().Get("1")
			require.Equal(t, vector.TypeArray, people.Type())
			require.Equal(t, 2, people.Limit())

			p0 := people.At(0)
			require.Equal(t, vector.TypeObject, p0.Type())
			require.Equal(t, "Alice", string(p0.Get("1").RawBytes()))
			require.Equal(t, int64(1234), mustNodeInt(t, p0.Get("2")))
			require.Equal(t, "alice@example.com", string(p0.Get("3").RawBytes()))

			phones := p0.Get("4")
			require.Equal(t, vector.TypeArray, phones.Type())
			require.Equal(t, 2, phones.Limit())
			require.Equal(t, "+1-555-0100", string(phones.At(0).Get("1").RawBytes()))
			require.Equal(t, vector.TypeNull, phones.At(0).Get("2").Type())
			require.Equal(t, int64(2), mustNodeInt(t, phones.At(1).Get("2")))

			ts0 := p0.Get("5")
			require.Equal(t, vector.TypeObject, ts0.Type())
			require.Equal(t, int64(1700000000), mustNodeInt(t, ts0.Get("1")))
			require.Equal(t, vector.TypeNull, ts0.Get("2").Type())

			p1 := people.At(1)
			require.Equal(t, vector.TypeObject, p1.Type())
			require.Equal(t, "Bob", string(p1.Get("1").RawBytes()))
			require.Equal(t, vector.TypeNull, p1.Get("4").Type())
			require.Equal(t, int64(456000000), mustNodeInt(t, p1.Get("5").Get("2")))
		},
	},
	{
		name: "ClickEvent",
		build: func() []byte {
			b, err := os.ReadFile("testdata/event.bin")
			if err != nil {
				panic(err)
			}
			return b
		},
		check: func(t *testing.T, vec *Vector) {
			require.Equal(t, vector.TypeObject, vec.Root().Type())
			require.Equal(t, 44, vec.Root().Limit())

			require.Equal(t, wireVarint, WireType(vec.Root().Get("1")))
			require.Equal(t, int64(1786378544), mustVecInt(t, vec, "1"))
			require.Equal(t, int64(1786207577), mustVecInt(t, vec, "2"))
			require.Equal(t, wireLen, WireType(vec.Root().Get("3")))
			require.Equal(t, "loot-888.store", mustVecString(t, vec, "3"))
			require.Equal(t, "46e81ba1-eb94-440f-8e04-279ad0a798fb", mustVecString(t, vec, "4"))
			require.Equal(t, "LAND_VIEW", mustVecString(t, vec, "5"))
			require.Equal(t, "USD", mustVecString(t, vec, "9"))
			require.Equal(t, "2800:810:50c:777:e19d:d851:a5b2:8519", mustVecString(t, vec, "11"))
			require.Equal(t, wireVarint, WireType(vec.Root().Get("12")))
			require.Equal(t, int64(27747), mustVecInt(t, vec, "12"))
			require.Equal(t, "Telecentro S.A.", mustVecString(t, vec, "13"))
			require.Equal(t, "AR", mustVecString(t, vec, "14"))
			require.Equal(t, "mobile", mustVecString(t, vec, "17"))
			require.Equal(t, "Android", mustVecString(t, vec, "18"))
			require.Equal(t, "14", mustVecString(t, vec, "20"))
			require.Equal(t, "Facebook", mustVecString(t, vec, "21"))
			require.Equal(t, "573.0.0.37.74", mustVecString(t, vec, "22"))
			require.Equal(t, "ES", mustVecString(t, vec, "23"))
			require.True(t, strings.Contains(mustVecString(t, vec, "25"), "Chrome/151.0.7922.102"))
			require.Equal(t, "-", mustVecString(t, vec, "27"))
			require.Equal(t, "-", mustVecString(t, vec, "28"))
			require.Equal(t, "cpa", mustVecString(t, vec, "40"))
			require.Equal(t, "facebook", mustVecString(t, vec, "47"))
			require.Equal(t, "{account.id}", mustVecString(t, vec, "48"))
			require.Equal(t, "Facebook_Mobile_Feed", mustVecString(t, vec, "49"))
			require.Equal(t, "120251338185160584", mustVecString(t, vec, "50"))
			require.Equal(t, "fb", mustVecString(t, vec, "53"))
			require.Equal(t, "08.08 Granpampa AR / Bonanza", mustVecString(t, vec, "55"))
			require.Equal(t, "1", mustVecString(t, vec, "56"))
			require.Equal(t, "1", mustVecString(t, vec, "57"))
			require.True(t, strings.HasPrefix(mustVecString(t, vec, "62"), "IwcGRvZgF"))
		},
	},
}

func TestExamples(t *testing.T) {
	for _, st := range stages {
		cpy := st
		t.Run(cpy.name, func(t *testing.T) {
			vec := NewVector()
			if err := vec.ParseCopy(cpy.build()); err != nil {
				t.Fatalf("parse: %v", err)
			}
			cpy.check(t, vec)
		})
	}
}

func BenchmarkExamples(b *testing.B) {
	for _, st := range stages {
		cpy := st
		src := cpy.build()
		b.Run(cpy.name, func(b *testing.B) {
			b.ReportAllocs()
			vec := NewVector()
			for i := 0; i < b.N; i++ {
				vec.Reset()
				if err := vec.Parse(src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
