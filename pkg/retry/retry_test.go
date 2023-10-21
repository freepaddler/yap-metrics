package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithStrategy(t *testing.T) {
	ctx := context.Background()
	errRecover := errors.New("recoverable")
	errNoRecover := errors.New("unrecoverable")
	checkErr := func(err error) bool {
		return errors.Is(err, errRecover)
	}
	type args struct {
		retries []int
		errs    []error
	}
	tests := []struct {
		name     string
		args     args
		wantErr  error
		wantRuns int
	}{
		{
			name:     "single run",
			wantErr:  nil,
			wantRuns: 1,
			args: args{
				retries: []int{1, 2, 3},
				errs:    []error{nil},
			},
		},
		{
			name:     "recover from 2 errors",
			wantErr:  nil,
			wantRuns: 3,
			args: args{
				retries: []int{1, 2, 3},
				errs:    []error{errRecover, errRecover, nil},
			},
		},
		{
			name:     "fail on unrecoverable error ",
			wantErr:  errNoRecover,
			wantRuns: 2,
			args: args{
				retries: []int{1, 2, 3},
				errs:    []error{errRecover, errNoRecover, nil},
			},
		},
		{
			name:     "fail on no more attempts",
			wantErr:  errRecover,
			wantRuns: 4,
			args: args{
				retries: []int{1, 2, 3},
				errs:    []error{errRecover, errRecover, errRecover, errRecover},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := 0
			startTime := time.Now()
			err := WithStrategy(
				ctx,
				func(ctx context.Context) error {
					err := func() (err error) {
						n++
						//fmt.Printf("run #%d\n", n)
						return tt.args.errs[n-1]
					}()
					return err
				},
				checkErr,
				tt.args.retries...,
			)
			stopTime := time.Now()
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.wantRuns, n, "Expected %d runs got %d", tt.wantRuns, n)
			expectedDur := func() time.Duration {
				s := 0
				if n > 1 {
					for i := 0; i < n-1; i++ {
						s += tt.args.retries[i]
					}
				}
				return time.Duration(s) * time.Second
			}()
			// add 1 second for any overhead
			assert.WithinRange(
				t,
				stopTime,
				startTime.Add(expectedDur),
				startTime.Add(expectedDur+time.Second),
				"Expected duration (1 second diff is ok) %.2f, got $.2f", expectedDur.Seconds()+1, stopTime.Sub(startTime).Seconds(),
			)
		})
	}
}
