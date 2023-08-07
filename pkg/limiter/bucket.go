package limiter

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

var ErrorEmptyBucket = errors.New("Bucket has no more tokens")

type Bucket struct {
	id       string
	owner    string
	count    int32
	capacity int32
}

func NewBucket(owner string, capacity int32) *Bucket {
	return &Bucket{id: uuid.New().String(), owner: owner, capacity: capacity, count: capacity}
}

func (bucket Bucket) IsEmpty() bool {
	return bucket.count == 0
}

func (bucket *Bucket) RemoveToken() error {
	if bucket.IsEmpty() {
		msg, _ := fmt.Printf("Bucket %s belonging to %s has no more tokens", bucket.id, bucket.owner)
		log.Println(msg)
		return ErrorEmptyBucket
	} else {
		bucket.count--
		return nil
	}
}

func (bucket *Bucket) RefillTokens() {
	bucket.count = bucket.capacity
}
