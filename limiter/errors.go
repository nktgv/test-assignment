package limiter

import "errors"

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
