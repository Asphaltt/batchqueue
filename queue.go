package batchqueue

// A queue is the local cache for userq.
type queue struct {
	// the values slice is for memory reusing.
	// When to reuse the local cache, the values slice will be
	// assigned as values[:0].
	values []interface{}

	// the deqIndex is the index of values slice for dequeueing.
	deqIndex int
}

// newQueue creates a new local cache with specified capacity.
func newQueue(capacity int) *queue {
	var q queue
	q.values = make([]interface{}, capacity)
	return &q
}

// Enqueue pushes a value to values slice.
func (q *queue) Enqueue(v interface{}) {
	q.values = append(q.values, v)
}

// Dequeue pops a value from values slice from front side.
func (q *queue) Dequeue() (v interface{}) {
	v = q.values[q.deqIndex]
	q.values[q.deqIndex] = nil
	q.deqIndex++
	return
}

// A Queue is a message queue with local cache. When its local cache
// is not full, it won't commit the caching values to the batchqueue.
// Or it can forcely commit the caching values by Flush.
//
// A Queue must be used by only one goroutine, because it enqueues
// or dequeues a value without locking.
//
// Using local cache for less locking.
type Queue interface {
	// Enqueue pushes a value to local cache. When the local cache
	// is full, it'll be commited to the batchqueue.
	Enqueue(v interface{})

	// Dequeue pops a value from local cache. When the local cache
	// is empty, it gets one local cache from batchqueue.
	Dequeue() (v interface{})

	// Flush forcely commits local cache to the batchqueue.
	Flush()
}

// An userq is a queue for every goroutine to enqueue or dequeue
// a value. It uses two local caches, one is for enqueueing, and the
// other is for dequeueing.
//
// When its local enqueueing cache is empty and to enqueue a value,
// it gets a local cache from the batchqueue.
// When after enqueueing a value, its local enqueueing cache becomes
// full, it commits the local cache to the batchqueue.
//
// When its local dequeueing cache is empty and to dequeue a value,
// it gets a local cache from the batchqueue.
// When after dequeueing a value, its local dequeueing cache becomes
// empty, it returns the local cache to the batchqueue.
type userq struct {
	enq, deq *queue
	b        *batch
}

// Enqueue pushes the value to local enqueueing cache.
func (u *userq) Enqueue(v interface{}) {
	if u.enq == nil {
		u.enq = u.b.get()
		u.enq.values = u.enq.values[:0]
	}
	u.enq.Enqueue(v)
	if len(u.enq.values) == cap(u.enq.values) {
		u.b.put(u.enq)
		u.enq = nil
	}
}

// Dequeue pops a value from local dequeueing cache.
func (u *userq) Dequeue() (v interface{}) {
	if u.deq == nil {
		u.deq = u.b.wait()
		u.deq.deqIndex = 0
	}
	v = u.deq.Dequeue()
	if u.deq.deqIndex == len(u.deq.values) {
		u.b.free(u.deq)
		u.deq = nil
	}
	return v
}

// Flush forcely commits its local enqueueing cache to the batchqueue,
// when and only when its local enqueueing cache is not empty.
func (u *userq) Flush() {
	if u.enq != nil && len(u.enq.values) != 0 {
		u.b.put(u.enq)
		u.enq = nil
	}
}
