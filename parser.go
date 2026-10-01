package pbvector

import "github.com/koykov/vector"

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
	fields, err := scanMessage(src, s, e)
	if err != nil {
		vec.SetErrOffset(s)
		return err
	}
	childrenDepth := depth + 1
	vec.NodeAt(parentIdx).SetOffset(vec.Index.Len(childrenDepth))

	order := make([]int, 0, len(fields))
	groups := make(map[int][]field, len(fields))
	for _, f := range fields {
		if _, seen := groups[f.num]; !seen {
			order = append(order, f.num)
		}
		groups[f.num] = append(groups[f.num], f)
	}

	for _, num := range order {
		occs := groups[num]
		if len(occs) == 1 {
			if err := vec.buildValue(parentIdx, childrenDepth, srcp, src, occs[0], rec); err != nil {
				return err
			}
			continue
		}
		arr, arrIdx := vec.AcquireNodeWithType(childrenDepth, vector.TypeArray)
		vec.setKey(arr, num)
		vec.NodeAt(parentIdx).SetLimit(vec.Index.Len(childrenDepth))
		vec.NodeAt(arrIdx).SetOffset(vec.Index.Len(childrenDepth + 1))
		for _, occ := range occs {
			if err := vec.buildValue(arrIdx, childrenDepth+1, srcp, src, occ, rec); err != nil {
				return err
			}
		}
		vec.NodeAt(arrIdx).SetLimit(vec.Index.Len(childrenDepth + 1))
	}
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
		if _, err := scanMessage(src, f.off, f.off+f.length); err == nil {
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
