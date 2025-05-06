package strategy

import (
	"errors"
	"sync"

	"http-load-balancer/models"
)

type LeastConnections struct {
	mu *sync.RWMutex
}

func NewLeastConnections(backends []models.Backend) (*LeastConnections, error) {
	if len(backends) == 0 {
		return &LeastConnections{}, errors.New("no URLs provided")
	}
	return &LeastConnections{
		mu: new(sync.RWMutex),
	}, nil
}

func (lc *LeastConnections) NextBackend(backends []models.Backend) (models.Backend, error) {
	var minConns = -1
	var selected models.Backend
	found := false

	lc.mu.Lock()
	for _, b := range backends {
		if !b.IsAlive {
			continue
		}
		if minConns == -1 || b.ActiveConns < minConns {
			minConns = b.ActiveConns
			selected = b
			found = true
		}
	}
	lc.mu.Unlock()

	if !found {
		return models.Backend{}, ErrNoAliveBackends
	}
	return selected, nil
}
