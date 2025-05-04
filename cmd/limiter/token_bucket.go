package limiter

import (
	"http-load-balancer/repository"
	"sync"
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
