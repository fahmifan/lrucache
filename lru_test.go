package lrucache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := NewQueue()
	{ // check InsertFirst
		for i := 0; i < 10; i++ {
			q.InsertFirst(&Node{Item: Item{
				Key:   fmt.Sprint(i),
				Value: fmt.Sprint(i),
			}})
		}

		node := q.head
		for i := 9; i > 0; i-- {
			if node == nil {
				break
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	}

	{ // check RemoveLast
		q.RemoveLast()
		node := q.head
		for i := 9; i > 1; i-- {
			if node == nil {
				break
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	}

	{ // check RemoveNode: head
		q.RemoveNode(q.head)
		node := q.head
		for i := 8; i > 1; i-- {
			if node == nil {
				break
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	}

	{ // check RemoveNode: tail
		q.RemoveNode(q.tail)
		node := q.head
		for i := 8; i > 2; i-- {
			if node == nil {
				break
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	}

	{ // check RemoveNode: middle node
		node := q.head
		for i := 8; i > 2; i-- {
			if i == 5 {
				assert.NotNil(t, node)
				break
			}
		}

		q.RemoveNode(node)
		for i := 8; i > 2; i-- {
			if node == nil {
				break
			}

			if i == 5 {
				// should not exists
				continue
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	}
}

func TestQueue_MoveFirst(t *testing.T) {
	t.Run("node is head", func(t *testing.T) {
		q := NewQueue()
		seeds(q, 1, 4)

		node := q.head
		q.MoveToFirst(node)

		for i := 4; i >= 1; i-- {
			if node == nil {
				break
			}

			assert.Equal(t, fmt.Sprint(i), node.Item.Value.(string))
			node = node.next
		}
	})

	t.Run("node is tail", func(t *testing.T) {
		q := NewQueue()
		seeds(q, 1, 4)
		node := q.tail
		q.MoveToFirst(node)

		assert.Equal(t, q.head, node)
		assert.Equal(t, "1", node.Item.Value.(string))

		node = node.next
		assert.Equal(t, "4", node.Item.Value.(string))

		node = node.next
		assert.Equal(t, "3", node.Item.Value.(string))

		node = node.next
		assert.Equal(t, "2", node.Item.Value.(string))
		assert.Equal(t, q.tail, node)
		assert.Nil(t, node.next)
	})

	t.Run("node is in middle", func(t *testing.T) {
		q := NewQueue()
		seeds(q, 1, 4)

		node := q.head
		assert.Equal(t, "4", node.Item.Value.(string))

		node = node.next
		node = node.next
		assert.Equal(t, "2", node.Item.Value.(string))

		q.MoveToFirst(node)
		assert.Equal(t, q.head, node)
		assert.Nil(t, node.prev)

		node = node.next
		assert.Equal(t, "4", node.Item.Value.(string))

		node = node.next
		assert.Equal(t, "3", node.Item.Value.(string))

		node = node.next
		assert.Equal(t, q.tail, node)
		assert.Equal(t, "1", q.tail.Item.Value.(string))
		assert.Nil(t, q.tail.next)
	})
}

func seeds(q *Queue, from, to int) {
	for i := 1; i <= 4; i++ {
		q.InsertFirst(&Node{Item: Item{
			Key:   fmt.Sprint(i),
			Value: fmt.Sprint(i),
		}})
	}

}

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
	F := fmt.Sprint
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Put(F(i), i)
		}(i)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Get(F(i % 50))
		}(i)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Del(F(i % 50))
		}(i)
	}

	wg.Wait()
}

func BenchmarkLRUCacher(b *testing.B) {
	b.Run("Put", func(b *testing.B) {
		lru := NewLRUCacher(1000)
		for i := 0; i < b.N; i++ {
			lru.Put(fmt.Sprint(i), i)
		}
	})

	lruSeeded := NewLRUCacher(1000)
	for i := 0; i < 1000; i++ {
		lruSeeded.Put(fmt.Sprint(i), i)
	}

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lruSeeded.Get(fmt.Sprint(i))
		}
	})

	b.Run("Del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// non sequential delete
			lruSeeded.Del(fmt.Sprint((i + 45) % 1000))
		}
	})
}
