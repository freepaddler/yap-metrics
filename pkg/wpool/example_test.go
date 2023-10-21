package wpool_test

import (
	"context"
	"fmt"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/pkg/wpool"
)

func Example() {
	logger.SetLevel("debug")
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
	// wait for tasks to complete
	pool.Stop(true)
	// new tasks can't be added
	err := pool.Task(func() { return })
	if err != nil {
		fmt.Println(err)
	}

	// Unordered output:
	// task 0 completed
	// task 1 completed
	// task 2 completed
	// task 3 completed
	// task 4 completed
	// task 5 completed
	// wpool is stopped
}
