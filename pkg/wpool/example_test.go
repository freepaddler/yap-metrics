package wpool_test

import (
	"context"
	"fmt"
	"time"

	"github.com/freepaddler/yap-metrics/pkg/wpool"
)

func Example() {
	// create and start new worker pool with two workers
	pool := wpool.New(context.Background(), 3)

	// run 6 tasks with 1 second wait
	for i := 0; i < 6; i++ {
		j := i
		err := pool.Task(func() {
			// simple func that prints
			time.Sleep(time.Second)
			fmt.Printf("task %d completed\n", j)
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	// stop pool and wait for tasks to complete
	<-pool.Stop()

	// Unordered output:
	// task 0 completed
	// task 1 completed
	// task 2 completed
	// task 3 completed
	// task 4 completed
	// task 5 completed
}
