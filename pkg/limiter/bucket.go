package limiter

import "errors"

type Bucket struct {
	Count    int32
	Capacity int32
}

var errBucketIsEmpty = errors.New("Bucket has no more tokens")

func NewBucket(capacity int32) *Bucket {
	return &Bucket{Count: capacity, Capacity: capacity}
}

func (bucket Bucket) IsEmpty() bool {
	return bucket.Count == 0
}

func (bucket *Bucket) RemoveToken() error {
	if bucket.IsEmpty() {
		return errBucketIsEmpty
	} else {
		bucket.Count--
		return nil
	}
}

func (bucket *Bucket) RefillTokens() {
	bucket.Count = bucket.Capacity
}
