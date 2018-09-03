package cmap

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Segment interface {
	// Put 会根据参数放入一个键-元素对
	// 第一个返回值表示是否新增了键-元素对
	Put(Pair) (bool, error)

	// Get 会根据给定参数返回对应的键-元素对
	// 该方法会根据根据给定的键计算元素值
	Get(key string) Pair

	// GetWithHash 会根据给定参数返回对应的键-元素对
	// 注意！参数 keyHash 是基于参数 key 计算得出的散列值
	GetWithHash(key string, keyHash uint64) Pair

	// 若返回值为 true 则说明已删除，否则说明没找到该键
	Delete(key string) bool

	// 获取当前段的尺寸(其中包括散列桶的数量)
	Size() uint64
}

func newSegment(bucketNum int, pairRedistributer PairRedistributor) Segment {
	if bucketNum <= 0 {
		bucketNum = DEFAULT_BUCKET_NUMBER
	}
	if pairRedistributer == nil {
		pairRedistributer = newDefaultPairRedistributor(
			DEFAULT_BUCKET_LOAD_FACTOR, bucketNum)
	}

	buckets := make([]Bucket, bucketNum)
	for i := 0; i < bucketNum; i++ {
		buckets[i] = newBucket()
	}

	return &segment{
		buckets:           buckets,
		bucketsLen:        bucketNum,
		pairRedistributor: pairRedistributer,
	}
}

type segment struct {
	// buckets 代表散列桶切片。
	buckets []Bucket
	// bucketsLen 代表散列桶切片的长度。
	bucketsLen int
	// pairTotal 代表本段里面键-元素对总数。
	pairTotal         uint64
	pairRedistributor PairRedistributor
	lock              sync.Mutex
}

func (s *segment) Put(p Pair) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	b := s.buckets[int(p.Hash()%uint64(s.bucketsLen))]
	ok, err := b.Put(p, nil)
	if ok {
		newTotal := atomic.AddUint64(&s.pairTotal, 1)
		s.redistribute(newTotal, b.Size())
	}
	return ok, err
}

func (s *segment) Get(key string) Pair {
	return s.GetWithHash(key, hash(key))
}

func (s *segment) GetWithHash(key string, keyHash uint64) Pair {
	s.lock.Lock()
	b := s.buckets[int(keyHash%uint64(s.bucketsLen))]
	s.lock.Unlock()

	return b.Get(key)
}

func (s *segment) Delete(key string) bool {
	s.lock.Lock()
	b := s.buckets[int(hash(key)%uint64(s.bucketsLen))]
	ok := b.Delete(key, nil)
	if ok {
		newTotal := atomic.AddUint64(&s.pairTotal, ^uint64(0))
		s.redistribute(newTotal, b.Size())
	}
	s.lock.Unlock()

	return ok
}

// 获取当前段的尺寸(其中包括散列桶的数量)
func (s *segment) Size() uint64 {
	return atomic.LoadUint64(&s.pairTotal)
}

func (s *segment) redistribute(pairTotal uint64, bucketSize uint64) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if pErr, ok := p.(error); ok {
				err = newPairRedistributorError(pErr.Error())
			} else {
				err = newPairRedistributorError(fmt.Sprintf("%s", p))
			}
		}
	}()

	s.pairRedistributor.UpdateThreshold(pairTotal, s.bucketsLen)
	bucketStatus := s.pairRedistributor.CheckBucketStatus(pairTotal, bucketSize)
	newBuckets, changed := s.pairRedistributor.Redistribute(bucketStatus, s.buckets)
	if changed {
		s.buckets = newBuckets
		s.bucketsLen = len(s.buckets)
	}
	return nil
}
