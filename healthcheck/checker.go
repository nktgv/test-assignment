package healthcheck

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"http-load-balancer/models"
	"http-load-balancer/repository"
)

type HealthChecker struct {
	repo       repository.BackendRepository
	httpClient *http.Client
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewHealthChecker(repo repository.BackendRepository, interval time.Duration) *HealthChecker {
	transport := &http.Transport{
		MaxIdleConns:          2000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	}
	return &HealthChecker{
		repo: repo,
		httpClient: &http.Client{
			Timeout:   time.Second,
			Transport: transport,
		},
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.run()
}

func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
}

func (hc *HealthChecker) run() {
	defer hc.wg.Done()

	hc.checkAllBackends()

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

	var wg sync.WaitGroup
	for _, b := range backends {
		wg.Add(1)
		go func(backend models.Backend) {
			defer wg.Done()
			hc.checkBackend(backend)
		}(b)
	}
	wg.Wait()
}

func (hc *HealthChecker) checkBackend(backend models.Backend) {
	url, err := url.Parse("http://" + backend.URL + "/health")
	if err != nil {
		return
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return
	}

	isAlive := true
	resp, err := hc.httpClient.Do(req)
	if err != nil {
		isAlive = false
	}
	defer resp.Body.Close()

	isAlive = isAlive && resp.StatusCode == http.StatusOK

	if _, err = hc.repo.SetIsAlive(backend.ID, isAlive); err != nil {
		return
	}
}
