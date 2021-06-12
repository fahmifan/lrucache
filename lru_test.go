package lrucache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUCacher_EmptySize(t *testing.T) {
	lru := &LRUCacher{}

	lru.Put("1", "1")
	lru.Put("2", "2")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))

	assert.Equal(t, DefaultMaxSize, lru.MaxSize)
}

func TestLRUCacher(t *testing.T) {
	lru := &LRUCacher{MaxSize: 3}

	lru.Put("1", "1")
	lru.Put("2", "2")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))

	lru.Put("3", "3")
	val = lru.Get("3")
	assert.Equal(t, "3", val.(string))

	// not found
	val = lru.Get("4")
	assert.Equal(t, nil, val)

	// already evicted
	lru.Put("4", "4")
	val = lru.Get("4")
	assert.Equal(t, "4", val.(string))
	val = lru.Get("1")
	assert.Equal(t, nil, val)

	val = lru.Del("4")
	assert.Equal(t, "4", val.(string))
	// deleted from the cache
	val = lru.Get("4")
	assert.Equal(t, nil, val)
}
