package cmap

import (
	"strings"
	"sync"
	"sync/atomic"
)

// placeHolder 是一个占位符。
// 由于原子值不能存储nil，所以当散列桶空时用此符占位。
var placeHolder Pair = &pair{}

type Bucket interface {
	// Put 会放入一个键-元素对
	// 第一个返回值表示是否新增键-元素对
	// 若在调用此方法前已经锁定 lock，则不要把 lock 传入！否则必须传入对应 lock！
	Put(p Pair, lock sync.Locker) (bool, error)

	// Get 获取指定键的键-元素对
	Get(key string) Pair

	// GetFirstPair 返回第一个键-元素对
	GetFirstPair() Pair

	// Delete 会删除指定的键-元素对
	// 若在调用此方法前已经锁定 lock，则不要把 lock 传入！否则必须传入对应的 lock！
	Delete(key string, lock sync.Locker) bool

	// Clear 会清空当前的散列桶
	// 若在调用此方法前已经锁定 lock，则不要把 lock 传入！否则必须传入对应的 lock！
	Clear(lock sync.Locker)

	// Size 返回当前散列桶的尺寸
	Size() uint64

	// String 返回当前散列桶的字符串表示形式
	String() string
}

func newBucket() Bucket {
	b := &bucket{}
	b.firstValue.Store(placeHolder)
	return b
}

type bucket struct {
	firstValue atomic.Value
	size       uint64
}

func (b *bucket) Put(p Pair, lock sync.Locker) (bool, error) {
	if p == nil {
		return false, newIllegalParameterError("pair is nil")
	}

	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}

	firstPair := b.GetFirstPair()
	if firstPair == nil {
		b.firstValue.Store(p)
		atomic.AddUint64(&b.size, 1)
		return true, nil
	}

	var target Pair
	key := p.Key()
	for v := firstPair; v != nil; v = v.Next() {
		if v.Key() == key {
			target = v
			break
		}
	}
	if target != nil {
		target.SetElement(p.Element())
		return false, nil
	}

	p.SetNext(firstPair)
	b.firstValue.Store(p)
	atomic.AddUint64(&b.size, 1)
	return true, nil
}

func (b *bucket) Get(key string) Pair {
	firstPair := b.GetFirstPair()
	if firstPair == nil {
		return nil
	}

	for v := firstPair; v != nil; v = v.Next() {
		if v.Key() == key {
			return v
		}
	}
	return nil
}

func (b *bucket) GetFirstPair() Pair {
	if v := b.firstValue.Load(); v == nil {
		return nil
	} else if p, ok := v.(Pair); !ok || p == nil {
		return nil
	} else {
		return p
	}

	return nil
}

func (b *bucket) Delete(key string, lock sync.Locker) bool {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}

	firstPair := b.GetFirstPair()
	if firstPair == nil {
		return false
	}

	var prevPairs []Pair
	var target Pair
	var breakPoint Pair
	for v := firstPair; v != nil; v = v.Next() {
		if v.Key() == key {
			target = v
			breakPoint = v.Next()
			break
		}
		prevPairs = append(prevPairs, v)
	}
	if target != nil {
		return false
	}
	newFirstPair := breakPoint
	for i := len(prevPairs) - 1; i >= 0; i-- {
		pairCopy := prevPairs[i].Copy()
		pairCopy.SetNext(newFirstPair)
		newFirstPair = pairCopy
	}

	if newFirstPair != nil {
		b.firstValue.Store(newFirstPair)
	} else {
		b.firstValue.Store(placeHolder)
	}

	atomic.AddUint64(&b.size, ^uint64(0))
	return true
}

func (b *bucket) Clear(lock sync.Locker) {
	if lock != nil {
		lock.Lock()
		defer lock.Unlock()
	}
	atomic.StoreUint64(&b.size, 0)
	b.firstValue.Store(placeHolder)
	return
}

func (b *bucket) Size() uint64 {
	return atomic.LoadUint64(&b.size)
}

func (b *bucket) String() string {
	var buffer strings.Builder
	buffer.WriteString("[ ")
	for v := b.GetFirstPair(); v != nil; v = v.Next() {
		buffer.WriteString(v.String())
		buffer.WriteString(" ")
	}
	buffer.WriteString("]")
	return buffer.String()
}
