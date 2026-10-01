package pbvector

func decodeVarintErr(b []byte) (uint64, int, error) {
	var v uint64
	for i := 0; i < len(b); i++ {
		if i >= 10 {
			return 0, 0, ErrVarintOverflow
		}
		c := b[i]
		if i == 9 && c > 1 {
			return 0, 0, ErrVarintOverflow
		}
		v |= uint64(c&0x7f) << (7 * uint(i))
		if c < 0x80 {
			return v, i + 1, nil
		}
	}
	return 0, 0, ErrTruncated
}

func decodeVarint(b []byte) (uint64, int, bool) {
	v, n, err := decodeVarintErr(b)
	return v, n, err == nil
}

type field struct {
	num    int
	wire   int
	off    int
	length int
}

func scanMessage(src []byte, s, e int) ([]field, error) {
	if s < 0 || e > len(src) || s > e {
		return nil, ErrTruncated
	}
	var fields []field
	for s < e {
		tag, n, err := decodeVarintErr(src[s:e])
		if err != nil {
			return nil, err
		}
		s += n
		num, wire := int(tag>>3), int(tag&7)
		if num < 1 {
			return nil, ErrUnexpectedWireType
		}
		var off, length int
		switch wire {
		case wireVarint:
			_, n2, err2 := decodeVarintErr(src[s:e])
			if err2 != nil {
				return nil, err2
			}
			off, length = s, n2
			s += n2
		case wireFixed64:
			if s+8 > e {
				return nil, ErrTruncated
			}
			off, length = s, 8
			s += 8
		case wireFixed32:
			if s+4 > e {
				return nil, ErrTruncated
			}
			off, length = s, 4
			s += 4
		case wireLen:
			l, n2, err2 := decodeVarintErr(src[s:e])
			if err2 != nil {
				return nil, err2
			}
			s += n2
			if l > uint64(e-s) {
				return nil, ErrTruncated
			}
			off, length = s, int(l)
			s += int(l)
		default:
			return nil, ErrUnexpectedWireType
		}
		fields = append(fields, field{num: num, wire: wire, off: off, length: length})
	}
	return fields, nil
}
