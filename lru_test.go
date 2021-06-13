package lrucache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUCacher_Del(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		lru := NewLRUCacher(0)
		val := lru.Del("empty")
		assert.Equal(t, nil, val)
	})

	t.Run("one", func(t *testing.T) {
		lru := NewLRUCacher(0)
		lru.Put("1", "1")
		assert.Equal(t, int64(1), lru.count)

		val := lru.Del("1")
		assert.Equal(t, "1", val.(string))
		assert.Equal(t, int64(0), lru.count)
	})

	t.Run("last item", func(t *testing.T) {
		lru := NewLRUCacher(0)
		lru.Put("1", "1")
		lru.Put("2", "2")

		val := lru.Del("1")
		assert.Equal(t, "1", val.(string))
	})

	t.Run("middle item", func(t *testing.T) {
		lru := NewLRUCacher(0)
		lru.Put("1", "1")
		lru.Put("2", "2")
		lru.Put("3", "3")

		val := lru.Del("2")
		assert.Equal(t, "2", val.(string))
	})
}

func TestLRUCacher_SizeOne(t *testing.T) {
	lru := NewLRUCacher(1)
	lru.Put("1", "1")
	lru.Put("2", "2")

	val := lru.Get("2")
	assert.Equal(t, "2", val.(string))

	val = lru.Get("1")
	assert.Equal(t, nil, val)
}

func TestLRUCacher_EmptySize(t *testing.T) {
	lru := NewLRUCacher(0)

	lru.Put("1", "1")
	lru.Put("2", "2")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))

	assert.Equal(t, DefaultMaxSize, lru.maxSize)
}

func TestLRUCacher(t *testing.T) {
	lru := NewLRUCacher(3)

	lru.Put("1", "1")
	lru.Put("2", "2")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))
	assert.Equal(t, 2, len(lru.hash))

	lru.Put("3", "3")
	val = lru.Get("3")
	assert.Equal(t, "3", val.(string))

	// not found
	val = lru.Get("4")
	assert.Equal(t, nil, val)
	assert.Equal(t, 3, len(lru.hash))

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
	assert.Equal(t, 2, len(lru.hash))
}

func TestLRUCacher_PutExistingKey(t *testing.T) {
	lru := NewLRUCacher(3)

	lru.Put("1", "1")
	val := lru.Get("1")
	assert.Equal(t, "1", val.(string))

	lru.Put("1", "foobar")
	val = lru.Get("1")
	assert.Equal(t, "foobar", val.(string))
}

func TestLRUCacher_Concurrent(t *testing.T) {
	wg := sync.WaitGroup{}
	lru := NewLRUCacher(3)
	F := fmt.Sprintf
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Put(F("%d", i), i)
		}(i)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Get(F("%d", i%50))
		}(i)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Del(F("%d", i%50))
		}(i)
	}

	wg.Wait()
}

func BenchmarkLRUCacher(b *testing.B) {
	b.Run("Put", func(b *testing.B) {
		lru := NewLRUCacher(1000)
		for i := 0; i < b.N; i++ {
			lru.Put(fmt.Sprintf("%d", i), i)
		}
	})

	lruSeeded := NewLRUCacher(1000)
	for i := 0; i < 1000; i++ {
		lruSeeded.Put(fmt.Sprintf("%d", i), i)
	}

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lruSeeded.Get(fmt.Sprintf("%d", i))
		}
	})

	b.Run("Del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// non sequential delete
			lruSeeded.Del(fmt.Sprintf("%d", (i+45)%1000))
		}
	})
}
