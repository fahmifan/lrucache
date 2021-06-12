package lrucache

import (
	"sync"
	"sync/atomic"
)

type Node struct {
	item  Item
	mutex sync.Mutex
	next  *Node
	prev  *Node
}

func (n *Node) SetItem(i Item) {
	n.mutex.Lock()
	n.item = i
	n.mutex.Unlock()
}

// set next & prev to nil
func (n *Node) breakLinks() {
	if n == nil {
		return
	}

	n.next = nil
	n.prev = nil
}

// Queue implemented in linked list
type Queue struct {
	head  *Node
	tail  *Node
	mutex sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		head: nil,
		tail: nil,
	}
}

func (q *Queue) isEmpty() bool {
	return q.head == nil && q.tail == nil
}

func (q *Queue) isOne() bool {
	return q.head != nil && q.head.next == nil
}

// Inert Node to the first of the queue
func (q *Queue) InsertFirst(newHead *Node) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.isEmpty() {
		q.head = newHead
		q.tail = newHead
		return
	}

	oldHead := q.head
	newHead.next = oldHead
	oldHead.prev = newHead
	q.head = newHead
}

// RemoveBack remove the last Node in the queue
func (q *Queue) RemoveLast() *Node {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.isEmpty() {
		return nil
	}

	if q.tail == nil {
		return nil
	}

	if q.isOne() {
		last := q.tail
		q.tail = nil
		q.head = nil
		last.breakLinks()
		return last
	}

	oldLast := q.tail
	newLast := q.tail.prev
	q.tail = newLast
	oldLast.breakLinks()
	return oldLast
}

// RemoveNode remove node from the queue
func (q *Queue) RemoveNode(node *Node) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.isEmpty() {
		return
	}

	if q.isOne() {
		q.head.breakLinks()
		q.tail.breakLinks()
		node.breakLinks()
		return
	}

	// node is first in the queue with following N-nodes
	if node == q.head {
		// new head is the next in the queue
		q.head = node.next
		node.breakLinks()
		return
	}

	// node is the last in the queue with previos N-nodes
	if node == q.tail {
		// new tail is the one before the node
		q.tail = node.prev
		node.breakLinks()
		return
	}

	// node is in the middle of the queue
	after := node.next
	before := node.prev
	// link the before & after
	before.next = after
	after.prev = before
	node.breakLinks()
}

type Item struct {
	Key   string
	Value interface{}
}

const DefaultMaxSize int64 = 24

// LRUCacher not concurrent safe
type LRUCacher struct {
	MaxSize int64

	queue       *Queue
	currentSize int64
	mutex       sync.Mutex

	hash      map[string]*Node
	hashMutex sync.RWMutex
}

func (l *LRUCacher) getItem(key string) *Node {
	l.hashMutex.RLock()
	node := l.hash[key]
	l.hashMutex.RUnlock()
	return node
}

func (l *LRUCacher) removeItem(item Item) {
	l.hashMutex.Lock()
	delete(l.hash, item.Key)
	l.hashMutex.Unlock()
}

func (l *LRUCacher) putItem(node *Node) {
	l.hashMutex.Lock()
	l.hash[node.item.Key] = node
	l.hashMutex.Unlock()

}

func (l *LRUCacher) queueIsFull() bool {
	return l.currentSize == l.MaxSize
}

func (l *LRUCacher) Put(key string, value interface{}) {
	if l.MaxSize < 1 {
		l.MaxSize = DefaultMaxSize
	}

	if l.queue == nil {
		l.mutex.Lock()
		l.queue = NewQueue()
		l.mutex.Unlock()
	}

	if l.hash == nil {
		l.hashMutex.Lock()
		l.hash = make(map[string]*Node)
		l.hashMutex.Unlock()
	}

	item := Item{Key: key, Value: value}

	// if key already exist just replace the cache item
	oldNode := l.getItem(key)
	if oldNode != nil {
		oldNode.SetItem(item)
		return
	}

	node := &Node{item: item}
	if l.queueIsFull() {
		last := l.queue.RemoveLast()
		if last == nil {
			return
		}
		l.removeItem(last.item)
		l.putItem(node)
		l.queue.InsertFirst(node)
		return
	}

	l.putItem(node)
	l.queue.InsertFirst(node)
	atomic.AddInt64(&l.currentSize, 1)
}

func (l *LRUCacher) Get(key string) interface{} {
	if l.hash == nil {
		return nil
	}

	l.hashMutex.RLock()
	defer l.hashMutex.RUnlock()

	val, ok := l.hash[key]
	if !ok {
		return nil
	}

	return val.item.Value
}

func (l *LRUCacher) Del(key string) interface{} {
	node := l.getItem(key)
	if node == nil {
		return nil
	}

	l.queue.RemoveNode(node)
	l.removeItem(node.item)
	return node.item.Value
}
