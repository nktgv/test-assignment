package strategy

import (
	"sync/atomic"

	"http-load-balancer/models"
)

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) NextBackend(backends []models.Backend) (models.Backend, error) {
	if len(backends) == 0 {
		return models.Backend{}, nil
	}

	activeBackends := make([]models.Backend, 0)
	for _, backend := range backends {
		if backend.IsAlive {
			activeBackends = append(activeBackends, backend)
		}
	}

	if len(activeBackends) == 0 {
		return models.Backend{}, nil
	}

	idx := atomic.AddUint64(&rr.counter, 1) % uint64(len(activeBackends))
	return activeBackends[idx], nil
}
