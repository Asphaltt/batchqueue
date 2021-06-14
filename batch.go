package batchqueue

import (
	"sync"

	"github.com/Asphaltt/hqu"
)

// A Batch is an in-memory concurrency-safe message queue.
//
// It manipulates local cache for enqueueing and dequeueing.
type Batch interface {
	GetQueue() Queue
}

// A batch is the in-memory concurrency-safe message queue,
// with working queues and freelist queues.
type batch struct {
	// A `freelist` is a concurrency-safe stack for recycling
	// local caches.
	freelist hqu.Stack

	// A `workingq` is a concurrency-safe queue for queueing
	// local caches.
	// The `wcond` is a sync.Cond for waiting a local cache
	// when current `workingq` is empty.
	workingq hqu.Queue
	wcond    *sync.Cond

	// The `capacity` is the capacity of local cache.
	capacity int
}

// NewBatch creates a batchqueue for `Batch`, with local cache
// capacity.
func NewBatch(capacity int) Batch {
	var b batch
	b.wcond = sync.NewCond(&b.workingq.Mutex)
	b.capacity = capacity
	return &b
}

// GetQueue creates a new `Queue` instance for a goroutine.
//
// For example:
//
//    func consuming(q Queue) {
//        for v := q.Dequeue() {
//            ...
//        }
//    }
//
//    func producing(q Queue) {
//        for i:=0; i<100; i++ {
//            q.Enqueue(i)
//        }
//        q.Flush() // flushes the left caching values
//    }
//
//    func main() {
//        b := NewBatch()
//        go consuming(b.GetQueue())
// 	      go producing(b.GetQueue())
//        ...
//    }
func (b *batch) GetQueue() Queue {
	var q userq
	q.b = b
	return &q
}

// get pops a local cache from `freelist`, or creates a new one
// when the `freelist` is empty.
func (b *batch) get() (q *queue) {
	v, ok := b.freelist.Pop()
	if ok {
		q = v.(*queue)
	} else {
		q = newQueue(b.capacity)
	}
	return
}

// free pushes a local cache to `freelist`.
func (b *batch) free(q *queue) {
	b.freelist.Push(q)
}

// put pushes a local cache to `workingq`.
func (b *batch) put(q *queue) {
	n := b.workingq.Enqueue1(q)
	if n == 1 {
		b.wcond.Signal()
	}
}

// wait pops a local cache from `workingq`, or waits for one
// when the `workingq` is empty.
func (b *batch) wait() (q *queue) {
	b.workingq.Lock()
	for {
		v, ok := b.workingq.Dequeue0()
		if !ok {
			b.wcond.Wait()
			continue
		}

		b.workingq.Unlock()
		q = v.(*queue)
		return
	}
}
