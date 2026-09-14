Go Concurrency Tasks
Task 1.1: Basic Background Goroutines (Fire-and-Forget)

Goal: Learn how to offload work out of the main request path using the go keyword.

What to Build

Create an endpoint:

POST /log


The endpoint receives a log payload.

The Task
Accept the request.
Immediately respond with 202 Accepted.
Spawn a goroutine to write the log to a file or simulate a 2-second disk write.
Key Learning

Verify that your API returns instantly without waiting for the background operation to complete.

Task 1.2: Waiting for Multiple Goroutines (sync.WaitGroup)

Goal: Learn to wait for parallel tasks to finish before returning a response.

What to Build

Create an endpoint:

POST /fetch-batch


The endpoint accepts a JSON array containing 5 URLs.

The Task
Initialize a sync.WaitGroup.
For each URL:
Call wg.Add(1).
Launch a goroutine.
Inside each goroutine:
Fetch the URL.
Call wg.Done() when finished.
Call wg.Wait() before returning the API response.
Key Learning
Fix the classic loop-variable capture bug inside goroutines.
Ensure all background tasks finish cleanly.
Understand how sync.WaitGroup coordinates multiple goroutines.
Task 1.3: Data Hand-off via Unbuffered Channels

Goal: Safely pass data out of background goroutines without shared memory or mutexes.

What to Build

Enhance the batch fetch endpoint:

POST /fetch-channel

The Task

Create an unbuffered channel:

ch := make(chan Result)


Instead of directly appending results to a shared slice—which creates a data race—have each goroutine send its result to ch.

Use a separate consumer goroutine or loop to read from ch until all expected results arrive.

Key Learning

Discover how unbuffered channels enforce synchronous hand-offs:

The sender blocks until the receiver reads the value.

Task 1.4: Queueing & Buffering (Buffered Channels)

Goal: Prevent request dropping during high load by implementing an in-memory queue.

What to Build

Create an endpoint:

POST /enqueue


and a background worker.

The Task

Create a buffered channel with a fixed capacity:

tasks := make(chan Task, 10)


The endpoint pushes incoming tasks directly into the channel without waiting for execution.

Create a persistent background worker goroutine that:

Reads tasks from the channel.
Processes tasks sequentially.

Handle channel saturation:

If the buffer is full, return:
503 Service Unavailable


Use a non-blocking channel send with select and default:

select {
case tasks <- task:
    // Task queued successfully
default:
    // Queue is full
}

Key Learning

Understand:

Non-blocking channel sends using select + default.
Buffered channel behavior.
In-memory queueing.
Handling backpressure under high load.
How worker goroutines process queued tasks.