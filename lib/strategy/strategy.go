package strategy

import (
	"errors"

	"http-load-balancer/models"
)

type Strategy interface {
	NextBackend(backends []models.Backend) (models.Backend, error)
}

var (
	ErrNoAliveBackends = errors.New("no alive backends")
)
