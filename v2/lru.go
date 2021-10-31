package lrucache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

// Item ..
type Item struct {
	mutex   sync.Mutex
	key     string
	value   interface{}
	element *list.Element
	// unix timestamp in second, <= 0 means no expiry
	expireAt int64
}

// NewItem ..
func NewItem(key string, val interface{}, expireAt int64) *Item {
	return &Item{
		key:      key,
		value:    val,
		expireAt: expireAt,
	}
}

func (n *Item) CheckExpire(now int64) bool {
	if n.expireAt == 0 {
		return false
	}
	return now > n.expireAt
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

type Option func(l *LRUCacher)

func WithOnEvicted(onEvicted func(key string, val interface{})) Option {
	return func(l *LRUCacher) {
		l.OnEvicted = onEvicted
	}
}

func WithAutoEviction(enable bool) Option {
	return func(l *LRUCacher) {
		l.enableAutoEviction = enable
	}
}

// LRUCacher ..
type LRUCacher struct {
	maxSize            int64
	enableAutoEviction bool

	list       *List
	count      int64
	countMutex sync.RWMutex

	hash      map[string]*Item
	hashMutex sync.RWMutex

	OnEvicted func(key string, val interface{})
}

func NewLRUCacher(maxSize int64, opts ...Option) *LRUCacher {
	if maxSize < 1 {
		maxSize = DefaultMaxSize
	}
	lru := &LRUCacher{
		list:    NewList(),
		hash:    make(map[string]*Item),
		maxSize: maxSize,
	}
	for _, opt := range opts {
		opt(lru)
	}
	if lru.enableAutoEviction {
		go lru.runJanitor()
	}
	return lru
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

type putOption struct {
	expireIn time.Duration // second
}

type PutOption func(o *putOption)

func WithExpireIn(expireIn time.Duration) PutOption {
	return func(o *putOption) {
		if o != nil {
			o.expireIn = expireIn
		}
	}
}

// Put ..
func (l *LRUCacher) Put(key string, value interface{}, opts ...PutOption) {
	// if key already exist just replace the cache item
	oldItem := l.getItem(key)
	if oldItem != nil {
		oldItem.SetValue(value)
		return
	}

	option := &putOption{}
	for _, opt := range opts {
		opt(option)
	}

	var expireAt int64
	if option.expireIn > 0 {
		expireAt = time.Now().Add(option.expireIn).Unix()
	}
	item := NewItem(key, value, expireAt)
	if l.queueIsFull() {
		last := l.list.RemoveBack()
		if last == nil {
			return
		}
		l.removeItem(last.key)
		item = l.list.PushFront(item)
		l.putItem(item)
		return
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

	if l.enableAutoEviction {
		now := time.Now().Unix()
		if item.CheckExpire(int64(now)) {
			return nil
		}
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

func (l *LRUCacher) Count() int64 {
	l.countMutex.RLock()
	count := l.count
	l.countMutex.RUnlock()
	return count
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

func (l *LRUCacher) runJanitor() {
	for {
		time.Sleep(2 * time.Second)
		tnow := time.Now()
		log.Println("janitor runs", tnow)

		l.hashMutex.RLock()
		var keysToEvict []string
		now := time.Now().Unix()
		for key, val := range l.hash {
			if val.CheckExpire(now) {
				keysToEvict = append(keysToEvict, key)
			}
		}
		l.hashMutex.RUnlock()

		for _, key := range keysToEvict {
			val := l.Del(key)
			if l.OnEvicted != nil {
				l.OnEvicted(key, val)
			}
		}

		log.Printf("janitor finished in %s\n", time.Since(tnow))
	}
}
