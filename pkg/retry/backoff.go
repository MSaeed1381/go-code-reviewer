package retry

import (
	"math/rand"
	"time"
)

func ExponentialBackoff(base time.Duration) Strategy {
	return func(attempt int) time.Duration {
		return base * (1 << (attempt - 1)) // base * 2^(attempt-1)
	}
}

func ExponentialJitterBackoff(base, max time.Duration) Strategy {
	return func(attempt int) time.Duration {
		backoff := base * (1 << (attempt - 1))
		if backoff > max {
			backoff = max
		}
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
		return backoff/2 + jitter
	}
}
