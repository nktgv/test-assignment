package limiter

import (
	"errors"
	"fmt"
	"http-load-balancer/repository"
	"sync"
	"time"
)

type TokenBucket struct {
	repo        repository.UserRepository
	defaultCap  int
	defaultRate int
	mu          sync.Mutex
}

func NewTokenBucket(repo repository.UserRepository, defaultCap, defaultRate int) *TokenBucket {
	return &TokenBucket{
		repo:        repo,
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
	}
}

func (tb *TokenBucket) Allow(userID uint64) (bool, error) {
	const op = "TokenBucket.Allow"
	tb.mu.Lock()
	defer tb.mu.Unlock()

	user, err := tb.repo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if errors.Is(err, repository.ErrUserNotFound) {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now()
	lastUpdated, _ := time.Parse(time.RFC3339, user.LastUpdated)
	elapsed := now.Sub(lastUpdated).Seconds()
	newTokens := int(elapsed * float64(user.RatePerSec))

	if newTokens > 0 {
		user.Tokens = min(user.Tokens+newTokens, user.Capacity)
		user.LastUpdated = now.Format(time.RFC3339)
		if _, err := tb.repo.UpdateTokens(user.ID, user.Tokens); err != nil {
			return false, err
		}
	}

	if user.Tokens <= 0 {
		return false, ErrRateLimitExceeded
	}

	user.Tokens--
	if _, err := tb.repo.UpdateTokens(user.ID, user.Tokens); err != nil {
		return false, err
	}

	return true, nil
}
