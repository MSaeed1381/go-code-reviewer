package retry

import (
	"context"
	"time"
)

var (
	defaultRetries             = 3
	defaultStrategy            = ExponentialBackoff(500 * time.Millisecond)
	defaultShouldRetryFunction = func(err error) bool { return err != nil }
)

type Strategy func(attempt int) time.Duration

type Options struct {
	MaxRetries  int
	Strategy    Strategy
	ShouldRetry func(error) bool
}

type Retrier[T any] interface {
	Do(ctx context.Context, fn func() (T, error)) (T, error)
}

type retrier[T any] struct {
	opts Options
}

func New[T any](opts Options) Retrier[T] {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = defaultRetries
	}
	if opts.Strategy == nil {
		opts.Strategy = defaultStrategy
	}
	if opts.ShouldRetry == nil {
		opts.ShouldRetry = defaultShouldRetryFunction
	}
	return &retrier[T]{opts: opts}
}

func (r *retrier[T]) Do(ctx context.Context, fn func() (T, error)) (T, error) {
	var zero, resp T
	var err error

	for attempt := 1; attempt <= r.opts.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		resp, err = fn()
		if err == nil {
			return resp, nil
		}

		if !r.opts.ShouldRetry(err) {
			return zero, err
		}

		if attempt == r.opts.MaxRetries {
			break
		}

		select {
		case <-time.After(r.opts.Strategy(attempt)):
		case <-ctx.Done():
			return zero, ctx.Err()
		}
	}

	return zero, err
}
