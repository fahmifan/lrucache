package lrucache

import (
	"container/list"
	"sync"
)

// Item ..
type Item struct {
	mutex   sync.Mutex
	key     string
	value   interface{}
	element *list.Element
}

// NewItem ..
func NewItem(key string, val interface{}) *Item {
	return &Item{
		key:   key,
		value: val,
	}
}

func (n *Item) SetElement(el *list.Element) {
	n.element = el
}

// SetValue ..
func (n *Item) SetValue(v interface{}) {
	n.mutex.Lock()
	n.value = v
	n.element.Value = v
	n.mutex.Unlock()
}

// DefaultMaxSize ..
const DefaultMaxSize int64 = 24

// LRUCacher ..
type LRUCacher struct {
	maxSize int64

	list       *List
	count      int64
	countMutex sync.RWMutex

	hash      map[string]*Item
	hashMutex sync.RWMutex
}

func NewLRUCacher(maxSize int64) *LRUCacher {
	if maxSize < 1 {
		maxSize = DefaultMaxSize
	}
	return &LRUCacher{
		list:    NewList(),
		hash:    make(map[string]*Item),
		maxSize: maxSize,
	}
}

func (l *LRUCacher) getItem(key string) *Item {
	l.hashMutex.RLock()
	node := l.hash[key]
	l.hashMutex.RUnlock()
	return node
}

func (l *LRUCacher) removeItem(key string) {
	l.hashMutex.Lock()
	delete(l.hash, key)
	l.hashMutex.Unlock()
}

func (l *LRUCacher) putItem(node *Item) {
	l.hashMutex.Lock()
	l.hash[node.key] = node
	l.hashMutex.Unlock()

}

func (l *LRUCacher) queueIsFull() bool {
	l.countMutex.RLock()
	count := l.count
	l.countMutex.RUnlock()
	return count == l.maxSize
}

// Put ..
func (l *LRUCacher) Put(key string, value interface{}) {
	// if key already exist just replace the cache item
	oldItem := l.getItem(key)
	if oldItem != nil {
		oldItem.SetValue(value)
		return
	}

	item := NewItem(key, value)
	if l.queueIsFull() {
		last := l.list.RemoveBack()
		if last == nil {
			return
		}
		l.removeItem(last.key)
	}

	item = l.list.PushFront(item)
	l.putItem(item)
	l.incCount()
}

// Get item by key and move the item to the first of the queue
func (l *LRUCacher) Get(key string) interface{} {
	l.hashMutex.RLock()
	defer l.hashMutex.RUnlock()

	if l.hash == nil {
		return nil
	}

	item, ok := l.hash[key]
	if !ok {
		return nil
	}

	l.list.MoveToFront(item)

	return item.value
}

// Del ..
func (l *LRUCacher) Del(key string) interface{} {
	node := l.getItem(key)
	if node == nil {
		return nil
	}

	l.list.Remove(node.element)
	l.removeItem(key)
	l.decCount()
	return node.value
}

func (l *LRUCacher) decCount() {
	l.countMutex.Lock()
	l.count--
	l.countMutex.Unlock()
}

func (l *LRUCacher) incCount() {
	l.countMutex.Lock()
	l.count++
	l.countMutex.Unlock()
}
