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
