package lrucache

import (
	"sync"
)

// Node ..
type Node struct {
	Item  Item
	mutex sync.Mutex
	next  *Node
	prev  *Node
}

// SetItem ..
func (n *Node) SetItem(i Item) {
	n.mutex.Lock()
	n.Item = i
	n.mutex.Unlock()
}

// set next & prev to nil
func (n *Node) breakLinks() {
	if n == nil {
		return
	}

	n.mutex.Lock()
	n.next = nil
	n.prev = nil
	n.mutex.Unlock()
}

// Queue implemented in linked list
type Queue struct {
	head  *Node
	tail  *Node
	mutex sync.Mutex
}

// NewQueue ..
func NewQueue() *Queue {
	return &Queue{
		head: nil,
		tail: nil,
	}
}

func (q *Queue) isEmpty() bool {
	return q.head == nil || q.tail == nil
}

func (q *Queue) isOne() bool {
	return q.head == q.head.next
}

// InsertFirst insert Node to the first of the queue
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

	// node is the last in the queue with previous N-nodes
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

// MoveToFirst move node to the first of the queue
func (q *Queue) MoveToFirst(node *Node) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// no need to move, there one or none in the queue
	if q.isEmpty() || q.isOne() {
		return
	}

	if q.head == node {
		return
	}

	if q.tail == node {
		beforeTail := node.prev
		q.tail = beforeTail
		beforeTail.next = nil

		node.breakLinks()
		node.next = q.head
		q.head = node
		return
	}

	nodeBefore := node.prev
	nodeAfter := node.next
	nodeBefore.next = nodeAfter
	nodeAfter.prev = nodeBefore
	node.breakLinks()
	node.next = q.head
	q.head = node
}

// Item ..
type Item struct {
	Key   string
	Value interface{}
}

// DefaultMaxSize ..
const DefaultMaxSize int64 = 24

// LRUCacher ..
type LRUCacher struct {
	maxSize int64

	queue      *Queue
	count      int64
	countMutex sync.RWMutex

	hash      map[string]*Node
	hashMutex sync.RWMutex
}

func NewLRUCacher(maxSize int64) *LRUCacher {
	if maxSize < 1 {
		maxSize = DefaultMaxSize
	}
	return &LRUCacher{
		queue:   NewQueue(),
		hash:    make(map[string]*Node),
		maxSize: maxSize,
	}
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
	l.hash[node.Item.Key] = node
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
	item := Item{Key: key, Value: value}

	// if key already exist just replace the cache item
	oldNode := l.getItem(key)
	if oldNode != nil {
		oldNode.SetItem(item)
		return
	}

	node := &Node{Item: item}
	if l.queueIsFull() {
		last := l.queue.RemoveLast()
		if last == nil {
			return
		}
		l.removeItem(last.Item)
		l.putItem(node)
		l.queue.InsertFirst(node)
		return
	}

	l.putItem(node)
	l.queue.InsertFirst(node)
	l.incCount()
}

// Get item by key and move the item to the first of the queue
func (l *LRUCacher) Get(key string) interface{} {
	l.hashMutex.RLock()
	defer l.hashMutex.RUnlock()

	if l.hash == nil {
		return nil
	}

	val, ok := l.hash[key]
	if !ok {
		return nil
	}

	l.queue.MoveToFirst(val)

	return val.Item.Value
}

// Del ..
func (l *LRUCacher) Del(key string) interface{} {
	node := l.getItem(key)
	if node == nil {
		return nil
	}

	l.queue.RemoveNode(node)
	l.removeItem(node.Item)
	l.decCount()
	return node.Item.Value
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
