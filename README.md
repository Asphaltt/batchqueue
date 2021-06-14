# batchqueue

[![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/Asphaltt/batchqueue) [![GitHub license](https://img.shields.io/github/license/Naereen/StrapDown.js.svg)](https://github.com/Naereen/StrapDown.js/blob/master/LICENSE)

A batchqueue is an in-memory concurrency-safe message queue by enqueueing and dequeueing a batch of messages.

A batchqueue shouldn't be used for unstable message-producing situations, like network packets. Because it won't commit the local enqueueing messages to the batchqueue when the local enqueueing cache hasn't been filled up. Or you can flush them with a timer.

## Enqueue

Enqueue pushes a value to its local cache.

When its local enqueueing cache is `nil`, it gets a local cache from batchqueue's `freelist`.

When its local enqueueing cache is filled up, it commits the local cache to batchqueue's `workingq`.

## Dequeue

Dequeue pops a value from its local cache.

When its local dequeueing cache is `nil`, it gets a local cache from batchqueue's `workingq`.

When its local dequeueing cache becomes empty, it returns the local cache to batchqueue's `freelist`.

## Usage

> go get github.com/Asphaltt/batchqueue

See `examples/pubsub.go`:

```go
package main

import (
	"fmt"
	"sync"

	bq "github.com/Asphaltt/batchqueue"
)

func main() {
	b := bq.NewBatch(8)
	var wg sync.WaitGroup
	wg.Add(128)
	// produce 128 messages
	producing(b.GetQueue(), 128)
	// start 8 goroutines to consume 128 messages
	for i := 0; i < 8; i++ {
		go func() {
			consuming(b.GetQueue(), 16)
			wg.Add(-16)
		}()
	}
	wg.Wait()
}

func consuming(q bq.Queue, n int) {
	for i := 0; i < n; i++ {
		v := q.Dequeue()
		fmt.Println(v)
	}
}

func producing(q bq.Queue, n int) {
	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}
	q.Flush()
}

```

