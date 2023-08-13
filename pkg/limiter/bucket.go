package limiter

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

type Bucket struct {
	Id       string `redis:"id"`
	Owner    string `redis:"owner"`
	Count    int32  `redis:"count"`
	Capacity int32  `redis:"capacity"`
}

var ErrBucketEmpty = errors.New("Bucket has no more tokens")

func NewBucket(owner string, capacity int32) *Bucket {
	return &Bucket{Id: uuid.New().String(), Owner: owner, Capacity: capacity, Count: capacity}
}

func (bucket Bucket) IsEmpty() bool {
	return bucket.Count == 0
}

func (bucket *Bucket) RemoveToken() error {
	if bucket.IsEmpty() {
		log.Printf("Bucket %s belonging to %s has no more tokens", bucket.Id, bucket.Owner)
		return ErrBucketEmpty
	} else {
		bucket.Count--
		return nil
	}
}

func (bucket *Bucket) RefillTokens() {
	bucket.Count = bucket.Capacity
}
