package lrucache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUCacher_Del(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		lru := &LRUCacher{}
		val := lru.Del("empty")
		assert.Equal(t, nil, val)
	})

	t.Run("one", func(t *testing.T) {
		lru := &LRUCacher{}
		lru.Put("1", "1")
		assert.Equal(t, 1, lru.count)

		val := lru.Del("1")
		assert.Equal(t, "1", val.(string))
		assert.Equal(t, 0, lru.count)
	})

	t.Run("last item", func(t *testing.T) {
		lru := &LRUCacher{}
		lru.Put("1", "1")
		lru.Put("2", "2")

		val := lru.Del("1")
		assert.Equal(t, "1", val.(string))
	})

	t.Run("middle item", func(t *testing.T) {
		lru := &LRUCacher{}
		lru.Put("1", "1")
		lru.Put("2", "2")
		lru.Put("3", "3")

		val := lru.Del("2")
		assert.Equal(t, "2", val.(string))
	})
}

func TestLRUCacher_SizeOne(t *testing.T) {
	lru := &LRUCacher{MaxSize: 1}
	lru.Put("1", "1")
	lru.Put("2", "2")

	val := lru.Get("2")
	assert.Equal(t, "2", val.(string))

	val = lru.Get("1")
	assert.Equal(t, nil, val)
}

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
	val = lru.Get("2")
	assert.Equal(t, nil, val)

	val = lru.Del("4")
	assert.Equal(t, "4", val.(string))
	// deleted from the cache
	val = lru.Get("4")
	assert.Equal(t, nil, val)
}

func TestLRUCacher_PutExistingKey(t *testing.T) {
	lru := &LRUCacher{MaxSize: 3}

	lru.Put("1", "1")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))

	lru.Put("1", "foobar")
	val = lru.Get("1")
	assert.Equal(t, "foobar", val.(string))
}
