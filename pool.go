package pbvector

import (
	"sync"

	"github.com/koykov/vector"
)

type Pool struct {
	p sync.Pool
}

var defaultPool Pool

func (p *Pool) Get() *Vector {
	if v := p.p.Get(); v != nil {
		if vec, ok := v.(*Vector); ok {
			vec.SetBit(vector.FlagInit, true)
			return vec
		}
	}
	return NewVector()
}

func (p *Pool) Put(vec *Vector) {
	vec.Reset()
	p.p.Put(vec)
}

func Acquire() *Vector { return defaultPool.Get() }

func Release(vec *Vector) { defaultPool.Put(vec) }

func ReleaseNC(vec *Vector) {
	vec.SetBit(vector.FlagNoClear, true)
	defaultPool.Put(vec)
}
