package batchqueue

import (
	"sync"
	"testing"
)

const defaultQueueCapacity = 8

func TestEnqueue(t *testing.T) {
	b := NewBatch(defaultQueueCapacity).(*batch)
	q := b.GetQueue().(*userq)

	reset := func() {
		b = NewBatch(defaultQueueCapacity).(*batch)
		q = b.GetQueue().(*userq)
	}

	tests := []struct {
		name   string
		run    func()
		expect func() bool
	}{
		{
			name: "empty",
			run:  func() {},
			expect: func() bool {
				return b.freelist.Size() == 0 && b.workingq.Size() == 0
			},
		},
		{
			name: "enqueue one message",
			run: func() {
				q.Enqueue(0)
			},
			expect: func() bool {
				return len(q.enq.values) == 1
			},
		},
		{
			name: "enqueue a batch of messages",
			run: func() {
				for i := 0; i < defaultQueueCapacity; i++ {
					q.Enqueue(i)
				}
				q.Flush()
			},
			expect: func() bool {
				return q.enq == nil &&
					b.workingq.Size() == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run()
			if !tt.expect() {
				t.Fail()
			}
		})
		reset()
	}
}

func TestDequeue(t *testing.T) {
	b := NewBatch(defaultQueueCapacity).(*batch)
	q := b.GetQueue().(*userq)

	reset := func() {
		b = NewBatch(defaultQueueCapacity).(*batch)
		q = b.GetQueue().(*userq)
	}

	tests := []struct {
		name   string
		run    func()
		expect func() bool
	}{
		{
			name: "empty",
			run:  func() {},
			expect: func() bool {
				return b.freelist.Size() == 0 && b.workingq.Size() == 0
			},
		},
		{
			name: "dequeue one message",
			run: func() {
				q.Enqueue(0)
				q.Flush()
			},
			expect: func() bool {
				v := q.Dequeue()
				return v == 0 &&
					b.freelist.Size() == 1
			},
		},
		{
			name: "dequeue a batch of messages",
			run: func() {
				for i := 0; i < defaultQueueCapacity; i++ {
					q.Enqueue(i)
				}
			},
			expect: func() bool {
				for i := 0; i < defaultQueueCapacity; i++ {
					v := q.Dequeue()
					if v != i {
						t.Logf("got %d, expect %d\n", v, i)
						return false
					}
				}
				return true
			},
		},
		{
			name: "dequeue more batch of messages",
			run: func() {
				for i := 0; i < defaultQueueCapacity*32; i++ {
					q.Enqueue(i)
				}
				q.Flush()
			},
			expect: func() bool {
				for i := 0; i < defaultQueueCapacity*32; i++ {
					v := q.Dequeue()
					if v != i {
						t.Logf("got %d, expect %d\n", v, i)
						return false
					}
				}
				return b.freelist.Size() == 32
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run()
			if !tt.expect() {
				t.Fail()
			}
		})
		reset()
	}
}

func TestFlush(t *testing.T) {
	b := NewBatch(defaultQueueCapacity).(*batch)
	q := b.GetQueue().(*userq)

	reset := func() {
		b = NewBatch(defaultQueueCapacity).(*batch)
		q = b.GetQueue().(*userq)
	}

	tests := []struct {
		name   string
		run    func()
		expect func() bool
	}{
		{
			name: "flush with no enqueueing",
			run:  func() { q.Flush() },
			expect: func() bool {
				return b.freelist.Size() == 0 && b.workingq.Size() == 0
			},
		},
		{
			name: "flush one message",
			run: func() {
				q.Enqueue(0)
				q.Flush()
			},
			expect: func() bool {
				return b.workingq.Size() == 1
			},
		},
		{
			name: "flush a batch of messages",
			run: func() {
				for i := 0; i < defaultQueueCapacity-1; i++ {
					q.Enqueue(i)
				}
				q.Flush()
			},
			expect: func() bool {
				return b.workingq.Size() == 1
			},
		},
		{
			name: "flush more batch of messages",
			run: func() {
				for i := 0; i < defaultQueueCapacity*32-1; i++ {
					q.Enqueue(i)
				}
				q.Flush()
			},
			expect: func() bool {
				return b.workingq.Size() == 32
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run()
			if !tt.expect() {
				t.Fail()
			}
		})
		reset()
	}
}

func TestPutAndWait(t *testing.T) {
	b := NewBatch(defaultQueueCapacity).(*batch)

	reset := func() {
		b = NewBatch(defaultQueueCapacity).(*batch)
	}

	tests := []struct {
		name   string
		run    func()
		expect func() bool
	}{
		{
			name: "do nothing",
			run:  func() {},
			expect: func() bool {
				return b.workingq.Size() == 0
			},
		},
		{
			name: "put one queue",
			run: func() {
				b.put(&queue{})
			},
			expect: func() bool {
				return b.workingq.Size() == 1
			},
		},
		{
			name: "put one queue and wait for it in sync way",
			run: func() {
				b.put(&queue{deqIndex: 99})
			},
			expect: func() bool {
				q := b.wait()
				return q.deqIndex == 99
			},
		},
		{
			name: "put one queue and wait for it in async way",
			run:  func() {},
			expect: func() bool {
				var q *queue
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					// make sure b.wait happens before b.put
					go func() {
						b.put(&queue{deqIndex: 99})
						wg.Done()
					}()
					q = b.wait()
					wg.Done()
				}()
				wg.Wait()
				return q.deqIndex == 99
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run()
			if !tt.expect() {
				t.Fail()
			}
		})
		reset()
	}
}
