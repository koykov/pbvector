package pbvector

import (
	"unicode"
	"unicode/utf8"

	"github.com/koykov/vector"
)

func (vec *Vector) parse(src []byte, copy bool) error {
	if !vec.CheckBit(vector.FlagInit) {
		return errBadInit
	}
	if err := vec.SetSrc(src, copy); err != nil {
		return err
	}
	src = vec.Src()
	srcp := vec.SrcAddr()

	_, rootIdx := vec.AcquireNodeWithType(0, vector.TypeObject)
	if err := vec.buildMessage(0, rootIdx, srcp, src, 0, len(src), 0); err != nil {
		return err
	}
	return nil
}

func (vec *Vector) buildMessage(depth, parentIdx int, srcp uintptr, src []byte, s, e, rec int) error {
	if rec > vec.maxDepth {
		return ErrRecursionLimit
	}
	buf := vec.acqbuf(depth)
	var err error
	buf.fields, err = scanMessage(buf.fields[:0], src, s, e)
	if err != nil {
		vec.SetErrOffset(s)
		vec.relbuf(depth, buf)
		return err
	}

	if buf.set == nil {
		buf.set = make(map[int]int, len(buf.fields))
	} else {
		clear(buf.set)
	}
	buf.order = buf.order[:0]
	buf.counts = buf.counts[:0]
	for _, f := range buf.fields {
		if i, ok := buf.set[f.num]; ok {
			buf.counts[i]++
		} else {
			buf.set[f.num] = len(buf.order)
			buf.order = append(buf.order, f.num)
			buf.counts = append(buf.counts, 1)
		}
	}

	n := len(buf.order)
	if cap(buf.starts) < n+1 {
		buf.starts = make([]int, n+1)
	} else {
		buf.starts = buf.starts[:n+1]
	}
	buf.starts[0] = 0
	for i := 0; i < n; i++ {
		buf.starts[i+1] = buf.starts[i] + buf.counts[i]
	}
	if cap(buf.cursor) < n {
		buf.cursor = make([]int, n)
	} else {
		buf.cursor = buf.cursor[:n]
	}
	copy(buf.cursor, buf.starts[:n])
	if cap(buf.gfields) < len(buf.fields) {
		buf.gfields = make([]field, len(buf.fields))
	} else {
		buf.gfields = buf.gfields[:len(buf.fields)]
	}
	for _, f := range buf.fields {
		i := buf.set[f.num]
		buf.gfields[buf.cursor[i]] = f
		buf.cursor[i]++
	}

	childrenDepth := depth + 1
	vec.NodeAt(parentIdx).SetOffset(vec.Index.Len(childrenDepth))

	for i, num := range buf.order {
		lo, hi := buf.starts[i], buf.starts[i+1]
		if hi-lo == 1 {
			if err := vec.buildValue(parentIdx, childrenDepth, srcp, src, buf.gfields[lo], rec); err != nil {
				vec.relbuf(depth, buf)
				return err
			}
			continue
		}
		arr, arrIdx := vec.AcquireNodeWithType(childrenDepth, vector.TypeArray)
		vec.setKey(arr, num)
		vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childrenDepth))
		vec.NodeAt(arrIdx).SetOffset(vec.Index.Len(childrenDepth + 1))
		for j := lo; j < hi; j++ {
			if err := vec.buildValue(arrIdx, childrenDepth+1, srcp, src, buf.gfields[j], rec); err != nil {
				vec.relbuf(depth, buf)
				return err
			}
		}
		vec.NodeAt(arrIdx).SetLimit(vec.Index.Len(childrenDepth + 1))
	}
	vec.relbuf(depth, buf)
	return nil
}

func (vec *Vector) buildValue(parentIdx, childDepth int, srcp uintptr, src []byte, f field, rec int) error {
	switch f.wire {
	case wireVarint, wireFixed64, wireFixed32:
		n, _ := vec.AcquireNodeWithType(childDepth, vector.TypeNumber)
		vec.setKey(n, f.num)
		n.Value().InitRaw(srcp, f.off, f.length)
		switch f.wire {
		case wireVarint:
			n.Value().SetBit(pbWireVarint, true)
		case wireFixed64:
			n.Value().SetBit(pbWireFixed64, true)
		case wireFixed32:
			n.Value().SetBit(pbWireFixed32, true)
		}
		vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childDepth))
	case wireLen:
		if f.length == 0 {
			n, _ := vec.AcquireNodeWithType(childDepth, vector.TypeObject)
			vec.setKey(n, f.num)
			vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childDepth))

			off := vec.Index.Len(childDepth + 1)
			if off == 0 {
				off = 1
			}
			n.SetOffset(off)
			n.SetLimit(off)
			return nil
		}
		if checktext(src[f.off : f.off+f.length]) {
			n, _ := vec.AcquireNodeWithType(childDepth, vector.TypeString)
			vec.setKey(n, f.num)
			n.Value().InitRaw(srcp, f.off, f.length)
			n.Value().SetBit(pbWireLen, true)
			vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childDepth))
			return nil
		}
		var err error
		cbuf := vec.acqbuf(childDepth)
		cbuf.fields, err = scanMessage(cbuf.fields[:0], src, f.off, f.off+f.length)
		vec.relbuf(childDepth, cbuf)
		if err == nil {
			n, idx := vec.AcquireNodeWithType(childDepth, vector.TypeObject)
			vec.setKey(n, f.num)
			vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childDepth))
			if err := vec.buildMessage(childDepth, idx, srcp, src, f.off, f.off+f.length, rec+1); err != nil {
				return err
			}
			return nil
		}
		n, _ := vec.AcquireNodeWithType(childDepth, vector.TypeString)
		vec.setKey(n, f.num)
		n.Value().InitRaw(srcp, f.off, f.length)
		n.Value().SetBit(pbWireLen, true)
		vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childDepth))
	}
	return nil
}

func (vec *Vector) setKey(n *vector.Node, num int) {
	kb := vec.putKey(num)
	n.Key().Init(kb, 0, len(kb))
}

func checktext(b []byte) bool {
	if len(b) == 0 || !utf8.Valid(b) {
		return false
	}
	for len(b) > 0 {
		r, sz := utf8.DecodeRune(b)
		if !unicode.IsPrint(r) {
			return false
		}
		b = b[sz:]
	}
	return true
}
