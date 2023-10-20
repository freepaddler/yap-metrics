package wpool_test

import (
	"context"
	"fmt"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/wpool"
)

func Example() {
	// create and start new worker pool with two workers
	pool := wpool.New(context.Background(), 2)

	// run 3 tasks
	for i := 0; i < 3; i++ {
		j := i
		pool.Task(func() {
			// simple func that prints
			time.Sleep(time.Second)
			fmt.Printf("task %d completed\n", j)
		})
	}

	// stop worker pool
	pool.Stop(context.Background(), true)

	// Unordered output:
	// task 0 completed
	// task 1 completed
	// task 2 completed
}
