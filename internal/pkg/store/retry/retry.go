package retry

import (
	"context"
	"errors"
	"net"
	"syscall"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// WithStrategy executes `try` function with the following retry strategy:
// If `try` execution result has error, this error is checked with isRetryError function.
// If resulted error requires retry, then next `try` execution delays on the interval (in seconds)
// from the next value of retries args. WithStrategy exits on the following conditions:
//
//	err==nil as a result of `try` execution
//	number of retries exceeded
//	`try` function error is not retryable (isRetryError(err)!=true)
//	context cancellation/expiration/timeout
func WithStrategy(
	ctx context.Context,
	try func(ctx context.Context) error,
	isRetryError func(error) bool,
	retries ...int) (err error) {

	for i := 0; i <= len(retries); i++ {
		logger.Log.Debug().Msgf("WithStrategy: try %d", i+1)
		err = try(ctx)
		if err == nil {
			logger.Log.Debug().Msg("WithStrategy: normal execution")
			return
		}
		if i == len(retries) {
			logger.Log.Debug().Msg("WithStrategy: no more tries left")
			return
		}
		if !isRetryError(err) {
			logger.Log.Warn().Err(err).Msg("WithStrategy: not retryable error")
			return
		}
		select {
		case <-time.After(time.Duration(retries[i]) * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return
}

// IsNetErr checks network timeout and connection refused errors
// may be used as isRetryError for WithStrategy function
func IsNetErr(err error) bool {
	if err == nil {
		return false
	}
	// connection refused
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	// network timeout
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}
