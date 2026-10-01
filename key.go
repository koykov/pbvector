package pbvector

import "strconv"

const keyChunkSize = 4096

func (vec *Vector) putKey(num int) []byte {
	var tmp [20]byte
	b := strconv.AppendInt(tmp[:0], int64(num), 10)
	if len(vec.keyChunks) == 0 {
		vec.keyChunks = append(vec.keyChunks, make([]byte, keyChunkSize))
	}
	chunk := vec.keyChunks[vec.keyIdx]
	if vec.keyOff+len(b) > len(chunk) {
		vec.keyIdx++
		vec.keyOff = 0
		if vec.keyIdx >= len(vec.keyChunks) {
			vec.keyChunks = append(vec.keyChunks, make([]byte, keyChunkSize))
		}
		chunk = vec.keyChunks[vec.keyIdx]
	}
	copy(chunk[vec.keyOff:], b)
	out := chunk[vec.keyOff : vec.keyOff+len(b)]
	vec.keyOff += len(b)
	return out
}
