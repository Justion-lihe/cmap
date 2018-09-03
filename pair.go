package cmap

import (
	"unsafe"
)

type linkedPair interface {
	Next() Pair
	SetNext(nextPair Pair) error
}

type Pair interface {
	linkedPair
	Hash() uint64
	Element() interface{}
	SetElement(interface{}) error
	Copy() Pair
	String() string
}

type pair struct {
	key     string
	hash    uint64
	element unsafe.Pointer
	next    unsafe.Pointer
}

func (p *pair) Next() Pair {
	return nil
}

func (p *pair) SetNext(nextPair Pair) error {
	return nil
}

func (p *pair) Hash() uint64 {
	return p.hash
}

func (p *pair) Element() interface{} {
	return nil
}

func (p *pair) SetElement(element interface{}) error {
	return nil
}

func (p *pair) Copy() Pair {
	return nil
}

func (p *pair) String() string {
	return ""
}

func newPair(key string, element interface{}) (Pair, error) {
	p := &pair{
		key:  key,
		hash: hash(key),
	}
	if element == nil {
		return nil, newIllegalParameterError("elemem")
	}
	p.element = unsafe.Pointer(&element)
	return nil, nil
}


