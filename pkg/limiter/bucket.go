package limiter

import (
	"errors"
	"log"
)

type Bucket struct {
	Owner    string `redis:"owner"`
	Count    int32  `redis:"count"`
	Capacity int32  `redis:"capacity"`
}

var ErrBucketEmpty = errors.New("Bucket has no more tokens")

func NewBucket(owner string, capacity int32) *Bucket {
	return &Bucket{Owner: owner, Capacity: capacity, Count: capacity}
}

func (bucket Bucket) IsEmpty() bool {
	return bucket.Count == 0
}

func (bucket *Bucket) RemoveToken() error {
	if bucket.IsEmpty() {
		log.Printf("Bucket belonging to %s has no more tokens", bucket.Owner)
		return ErrBucketEmpty
	} else {
		bucket.Count--
		return nil
	}
}

func (bucket *Bucket) RefillTokens() {
	bucket.Count = bucket.Capacity
}
