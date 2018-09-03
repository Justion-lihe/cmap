package cmap

type Segment interface {
	Put(Pair) (bool, error)
	GetWithHash(key string, keyHash uint64) Pair
	Delete(key string) bool
}

type segment struct {

}



func newSegment(bucketNum int, redistributer PairRedistributor) Segment {
	return nil
}

