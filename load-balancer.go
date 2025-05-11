package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	isAlive      bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.isAlive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.isAlive
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Site unreachable: %s", err)
		return false
	}
	defer conn.Close()

	return true
}

type LoadBalancer struct {
	backends []*Backend
	current  uint64
}

func (lb *LoadBalancer) NextBackend() *Backend {
	// round-robin
	next := int(atomic.AddUint64(&lb.current, uint64(1)) % uint64(len(lb.backends)))

	numberOfBackends := len(lb.backends)
	for i := range numberOfBackends {
		index := (next + i) % numberOfBackends
		if lb.backends[index].IsAlive() {
			return lb.backends[index]
		}
	}

	return nil
}

func (lb *LoadBalancer) HealthCheck() {
	for _, b := range lb.backends {
		status := isBackendAlive(b.URL)
		b.SetAlive(status)
		if status {
			log.Printf("Backend %s is alive", b.URL)
		} else {
			log.Printf("Backend %s is down", b.URL)
		}
	}
}

func (lb *LoadBalancer) HealthCheckPeriodically(interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-t.C:
			lb.HealthCheck()
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextBackend()
	if backend == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	backend.ReverseProxy.ServeHTTP(w, r)
}

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	flag.Parse()

	backendList := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	lb := LoadBalancer{}
	for _, serverUrl := range backendList {
		url, err := url.Parse(serverUrl)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Error: %v", err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		}

		lb.backends = append(lb.backends, &Backend{
			URL:          url,
			isAlive:      true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured backend: %s", url)
	}

	lb.HealthCheck()

	go lb.HealthCheckPeriodically(time.Minute)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: &lb,
	}

	log.Printf("Load Balancer started at :%d\n", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
