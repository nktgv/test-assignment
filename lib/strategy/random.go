package strategy

import (
	"fmt"
	"http-load-balancer/models"
	"math/rand"
	"time"
)

type Random struct {
	rnd *rand.Rand
}

func NewRandom(backends []models.Backend) *Random {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Random{
		rnd: rand.New(src),
	}
}

func (r *Random) NextBackend(backends []models.Backend) (models.Backend, error) {
	aliveBackends := make([]models.Backend, 0)
	for _, b := range backends {
		if b.IsAlive {
			aliveBackends = append(aliveBackends, b)
		}
	}

	if len(aliveBackends) == 0 {
		return models.Backend{}, fmt.Errorf("no alive backends")
	}

	return aliveBackends[r.rnd.Intn(len(aliveBackends))], nil
}
