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

	// Key 返回键的值
	Key() string

	// Hash 返回键的数列值
	Hash() uint64

	// Element 返回元素的值
	Element() interface{}

	// SetElement 设置元素的值
	SetElement(interface{}) error

	// Copy 生成一个当前键-值元素对的副本并返回
	Copy() Pair

	// String 返回当前键-值元素对的字符串表示形式
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

func (p *pair) Key() string {
	return p.key
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
