package cmap

import (
	"sync/atomic"
	"math"
)

type ConcurrentMap interface {
	// Concurrency 返回并发量
	Concurrency() int

	// Put 会推送一个键-元素对
	// 注意！参数 element 的值不能为 nil
	// 第一个返回值表示是否新增了元素对
	// 若键已存在，新元素会替换旧的元素值
	Put(key string, value interface{}) (bool, error)

	// Get 会获取与指定键关联的那个元素
	// 若返回 nil, 则说明指定的键不存在
	Get(key string) interface{}

	// Delete 会删除指定的键-元素对
	// 若结果值为 true, 则说明键已存在且已删除，否则说明键不存在
	Delete(key string) bool

	// Len 返回当前字典中键-元素对的数量
	Len() uint64
}

func NewConcurrentMap(concurrency int, pairRedistributor PairRedistributor) (ConcurrentMap, error) {
	if concurrency <= 0 {
		return nil, newIllegalParameterError("concurrency is too small")
	}
	if concurrency > MAX_CONCURRENCY {
		return nil, newIllegalParameterError("concurrency is too large")
	}

	cmap := &myConcurrentMap{}
	cmap.concurrency = concurrency
	cmap.segment = make([]Segment, concurrency)
	for i := 0; i < concurrency; i++ {
		cmap.segment[i] = newSegment(DEFAULT_BUCKET_NUMBER, pairRedistributor)
	}

	return nil, nil
}

type myConcurrentMap struct {
	concurrency int
	segment     []Segment
	total       uint64
}

func (m *myConcurrentMap) Concurrency() int {
	return m.concurrency
}

func (m *myConcurrentMap) Put(key string, element interface{}) (bool, error) {
	p, err := newPair(key, element)
	if err != nil {
		return false, err
	}

	s := m.findSegment(p.Hash())

	ok, err := s.Put(p)
	if ok {
		atomic.AddUint64(&m.total, 1)
	}
	return ok, err
}

func (m *myConcurrentMap) Get(key string) interface{} {
	keyHash := hash(key)
	s := m.findSegment(keyHash)
	pair := s.GetWithHash(key, keyHash)
	if pair == nil {
		return nil
	}
	return pair.Element()
}

func (m *myConcurrentMap) Delete(key string) bool {
	s := m.findSegment(hash(key))
	if s.Delete(key) {
		atomic.AddUint64(&m.total, ^uint64(1)+1)
		return true
	}
	return false
}

// Len 返回当前字典中键-元素对的数量
func (m *myConcurrentMap) Len() uint64 {
	return 0
}

func (m *myConcurrentMap) findSegment(keyHash uint64) Segment {
	if m.concurrency == 1 {
		return m.segment[0]
	}
	var keyHash16 uint16
	if keyHash > math.MaxUint16 {
		keyHash16 = uint16(keyHash >> 48)
	} else {
		keyHash16 = uint16(keyHash)
	}

	return m.segment[int(keyHash16)%m.concurrency]
}
