package signedcounter

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	key   []byte
	count uint64
}

func New(key []byte) (*Counter, error) {
	if len(key) < 16 {
		return nil, errors.New("key must contain at least 16 bytes")
	}
	return &Counter{key: append([]byte(nil), key...)}, nil
}
func (c *Counter) Sign(data []byte) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.key) == 0 {
		panic("signedcounter: use New")
	}
	c.count++
	m := hmac.New(sha256.New, c.key)
	m.Write(data)
	return m.Sum(nil)
}
