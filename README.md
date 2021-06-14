# batchqueue

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
