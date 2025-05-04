package strategy

import (
	"http-load-balancer/models"
)

type Strategy interface {
	NextBackend(backends []models.Backend) (models.Backend, error)
}
