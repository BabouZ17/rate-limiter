package limiter

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

var ErrBucketEmpty = errors.New("Bucket has no more tokens")

type Bucket struct {
	Id       string `redis:"id"`
	Owner    string `redis:"owner"`
	Count    int32  `redis:"count"`
	Capacity int32  `redis:"capacity"`
}

func NewBucket(owner string, capacity int32) *Bucket {
	return &Bucket{Id: uuid.New().String(), Owner: owner, Capacity: capacity, Count: capacity}
}

func (bucket Bucket) IsEmpty() bool {
	return bucket.Count == 0
}

func (bucket *Bucket) RemoveToken() error {
	if bucket.IsEmpty() {
		msg, _ := fmt.Printf("Bucket %s belonging to %s has no more tokens", bucket.Id, bucket.Owner)
		log.Println(msg)
		return ErrBucketEmpty
	} else {
		bucket.Count--
		return nil
	}
}

func (bucket *Bucket) RefillTokens() {
	bucket.Count = bucket.Capacity
}
