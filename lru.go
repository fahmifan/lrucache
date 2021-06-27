package lrucache

type Node struct {
	item Item
	next *Node
	prev *Node
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
	head *Node
	tail *Node
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
	if q.isEmpty() {
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

func (q *Queue) MoveToFirst(node *Node) {
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

type Item struct {
	Key   string
	Value interface{}
}

const DefaultMaxSize = 24

// LRUCacher not concurrent safe
type LRUCacher struct {
	queue   *Queue
	hash    map[string]*Node
	MaxSize int
	count   int
}

func (l *LRUCacher) removeItem(item Item) {
	delete(l.hash, item.Key)
}

func (l *LRUCacher) queueIsFull() bool {
	return l.count == l.MaxSize
}

func (l *LRUCacher) Put(key string, value interface{}) {
	if l.MaxSize < 1 {
		l.MaxSize = DefaultMaxSize
	}

	if l.queue == nil {
		l.queue = NewQueue()
	}

	if l.hash == nil {
		l.hash = make(map[string]*Node)
	}

	item := Item{
		Key:   key,
		Value: value,
	}

	// if key already exist just replace the cache item
	oldNode, ok := l.hash[key]
	if ok {
		oldNode.item = item
		return
	}

	node := &Node{item: item}
	if l.queueIsFull() {
		last := l.queue.RemoveLast()
		l.removeItem(last.item)

		l.hash[key] = node
		l.queue.InsertFirst(node)
		return
	}

	l.hash[key] = node
	l.queue.InsertFirst(node)
	l.count++
}

func (l *LRUCacher) Get(key string) interface{} {
	if l.hash == nil {
		return nil
	}

	val, ok := l.hash[key]
	if !ok {
		return nil
	}

	l.queue.MoveToFirst(val)

	return val.item.Value
}

func (l *LRUCacher) Del(key string) interface{} {
	node, ok := l.hash[key]
	if !ok {
		return nil
	}

	l.queue.RemoveNode(node)
	l.removeItem(node.item)
	l.count--
	return node.item.Value
}
