package healthcheck

import (
	"http-load-balancer/models"
	"http-load-balancer/repository"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	repo       repository.BackendRepository
	httpClient *http.Client
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewHealthChecker(repo repository.BackendRepository, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		repo: repo,
		httpClient: &http.Client{
			Timeout: time.Second * 5,
		},
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	hc.run()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
}

func (hc *HealthChecker) run() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllBackends()
		case <-hc.stopChan:
			return
		}
	}
}

func (hc *HealthChecker) checkAllBackends() {
	backends, err := hc.repo.GetAll()
	if err != nil {
		return
	}
	
	for _, b := range backends {
		go hc.checkBackend(b)
	}
}

func (hc *HealthChecker) checkBackend(backend models.Backend) {
	resp, err := hc.httpClient.Get(backend.Url)
	isAlive := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}
	if isAlive != backend.IsAlive {
		hc.repo.SetIsAlive(backend.ID, true)
	}
}
