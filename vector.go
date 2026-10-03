package pbvector

import (
	"errors"
	"io"
	"os"

	"github.com/koykov/byteconv"
	"github.com/koykov/vector"
)

const recursionLimit = 10000

const (
	pbWireVarint  = 4
	pbWireFixed64 = 5
	pbWireLen     = 6
	pbWireFixed32 = 7
)

const (
	wireVarint  = 0
	wireFixed64 = 1
	wireLen     = 2
	wireFixed32 = 5
)

var errBadInit = errors.New("pbvector: bad vector initialization, use pbvector.NewVector() or pbvector.Acquire()")

type Vector struct {
	vector.Vector
	maxDepth  int
	keyChunks [][]byte
	keyIdx    int
	keyOff    int
}

//go:noinline
func NewVector() *Vector {
	vec := &Vector{maxDepth: recursionLimit}
	vec.SetBit(vector.FlagInit, true)
	vec.SetCodec(Codec{})
	return vec
}

func (vec *Vector) Parse(src []byte) error { return vec.parse(src, false) }

func (vec *Vector) ParseCopy(src []byte) error { return vec.parse(src, true) }

func (vec *Vector) ParseString(src string) error { return vec.parse(byteconv.S2B(src), false) }

func (vec *Vector) ParseCopyString(src string) error { return vec.parse(byteconv.S2B(src), true) }

func (vec *Vector) ParseFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return vec.ParseReader(f)
}

func (vec *Vector) ParseReader(r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return vec.parse(buf, false)
}

func (vec *Vector) Reset() {
	vec.Vector.Reset()
	vec.keyIdx, vec.keyOff = 0, 0
	vec.SetBit(vector.FlagInit, true)
}
