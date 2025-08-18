package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errFail = errors.New("fail")

func TestSuccess_OnFirstTry(t *testing.T) {
	r := New[string](Options{MaxRetries: 3})

	called := 0
	resp, err := r.Do(context.Background(), func() (string, error) {
		called++
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
	assert.Equal(t, 1, called)
}

func TestEventuallySucceeds(t *testing.T) {
	r := New[string](Options{
		MaxRetries: 3,
		Strategy:   ExponentialBackoff(1 * time.Millisecond),
	})

	called := 0
	resp, err := r.Do(context.Background(), func() (string, error) {
		called++
		if called < 2 {
			return "", errFail
		}
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.Equal(t, 2, called)
}

func TestAlwaysFails(t *testing.T) {
	r := New[string](Options{
		MaxRetries: 3,
		Strategy:   ExponentialBackoff(1 * time.Millisecond),
	})

	called := 0
	resp, err := r.Do(context.Background(), func() (string, error) {
		called++
		return "", errFail
	})

	require.Error(t, err)
	assert.Equal(t, "", resp)
	assert.Equal(t, 3, called)
}

func TestShouldRetryFalse(t *testing.T) {
	r := New[string](Options{
		MaxRetries:  5,
		Strategy:    ExponentialBackoff(1 * time.Millisecond),
		ShouldRetry: func(err error) bool { return false },
	})

	called := 0
	resp, err := r.Do(context.Background(), func() (string, error) {
		called++
		return "", errFail
	})

	require.Error(t, err)
	assert.Equal(t, "", resp)
	assert.Equal(t, 1, called)
}

func TestContextCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New[string](Options{MaxRetries: 3})
	resp, err := r.Do(ctx, func() (string, error) {
		t.Fatal("function should not have been called")
		return "nope", nil
	})

	require.Error(t, err)
	assert.Equal(t, "", resp)
}

func TestContextCancelledDuringRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	r := New[string](Options{
		MaxRetries: 3,
		Strategy:   ExponentialBackoff(100 * time.Millisecond),
	})

	called := 0
	resp, err := r.Do(ctx, func() (string, error) {
		called++
		return "", errFail
	})

	require.Error(t, err)
	assert.Equal(t, "", resp)
	assert.GreaterOrEqual(t, called, 1)
}
