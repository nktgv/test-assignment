package strategy

import (
	"crypto/rand"
	"encoding/binary"
	"errors"

	"http-load-balancer/models"
)

type Random struct{}

func NewRandom() *Random {
	return &Random{}
}

func (r *Random) NextBackend(backends []models.Backend) (models.Backend, error) {
	aliveBackends := make([]models.Backend, 0, len(backends))
	for _, b := range backends {
		if b.IsAlive {
			aliveBackends = append(aliveBackends, b)
		}
	}

	// generate crypto-safety random number
	var n uint32
	err := binary.Read(rand.Reader, binary.BigEndian, &n)
	if err != nil {
		return models.Backend{}, err
	}
	if len(aliveBackends) == 0 {
		return models.Backend{}, errors.New("no alive backends")
	}

	selected := aliveBackends[int(n)%len(aliveBackends)]
	return selected, nil
}
