package cmap

type BucketStatus int

type PairRedistributor interface {
	// 根据键-元素对和散列桶总数计算并更新阈值
	UpdateThreshold(pairTotal uint64, bucketNumber int)

	// 检查散列桶的状态
	CheckBucketStatus(pairTotal uint64, bucketSize uint64) BucketStatus

	// 用于实施键-元素的再分布
	Redistribute(bucketStatus BucketStatus, buckets []Bucket) (newBuckets []Bucket, changed bool)
}

func newDefaultPairRedistributor(loadFactor float64, bucketNumber int) PairRedistributor {
	return nil
}
