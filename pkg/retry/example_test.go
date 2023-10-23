package retry_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/freepaddler/yap-metrics/pkg/retry"
)

func ExampleWithStrategy() {
	// test error
	testErr := errors.New("test error")

	// run count
	n := 0

	err := retry.WithStrategy(
		context.Background(),
		func(ctx context.Context) error {
			// body of wrapped function
			err := func() (err error) {
				n++
				fmt.Printf("run #%d\n", n)
				// return error on 1st and 2nd run
				if n < 3 {
					return testErr
				}
				return nil
			}()
			return err
		},
		// function to check is error is recoverable
		func(err error) bool {
			return errors.Is(err, testErr)
		},
		// timeouts for 3 retries
		1, 2, 3,
	)

	if err != nil {
		panic("retry failed")
	}

	// Output:
	// run #1
	// run #2
	// run #3
}
