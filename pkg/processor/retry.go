package processor

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type RetryableFunc func() error

func (p *processor) RetryWithBackoff(maxRetries int, fn RetryableFunc) error {
	for retry := 0; retry < maxRetries; retry++ {
		err := fn()
		if err == nil {
			return nil
		}

		backoffTime := time.Duration(1<<retry) * time.Second
		jitter := time.Duration(rand.Int63n(int64(backoffTime))) / 2
		sleepTime := backoffTime + jitter

		p.logger.Error(fmt.Sprintf("Error, retrying in %v, attempt %d/%d", sleepTime, retry+1, maxRetries), zap.Error(err))
		time.Sleep(sleepTime)
	}

	return &InternalError{
		Message: fmt.Sprintf("operation failed after max retries"),
	}
}
