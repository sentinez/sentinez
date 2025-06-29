package resolver

import (
	"errors"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type ServiceInstance struct {
	Name    string
	Address string
	Port    int
}

type ServiceDiscovery interface {
	Discover(name string) ([]ServiceInstance, error)
}

type Resolver struct {
	discovery ServiceDiscovery
	cache     map[string][]ServiceInstance
	lastFetch map[string]time.Time
	breakers  map[string]*gobreaker.CircuitBreaker
	mu        sync.RWMutex
	index     map[string]int
}

func New(sd ServiceDiscovery) *Resolver {
	return &Resolver{
		discovery: sd,
		cache:     make(map[string][]ServiceInstance),
		lastFetch: make(map[string]time.Time),
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		index:     make(map[string]int),
	}
}

func (r *Resolver) GetInstances(name string) ([]ServiceInstance, error) {
	r.mu.RLock()
	instances, ok := r.cache[name]
	last := r.lastFetch[name]
	r.mu.RUnlock()

	if ok && time.Since(last) < 10*time.Second {
		return instances, nil
	}

	newInstances, err := r.discovery.Discover(name)
	if err != nil || len(newInstances) == 0 {
		return nil, errors.New("no service available")
	}

	r.mu.Lock()
	r.cache[name] = newInstances
	r.lastFetch[name] = time.Now()
	r.mu.Unlock()

	return newInstances, nil
}

func (r *Resolver) PickInstance(name string) (*ServiceInstance, error) {
	instances, err := r.GetInstances(name)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(instances) == 0 {
		return nil, errors.New("no instance available")
	}

	idx := r.index[name] % len(instances)
	r.index[name] = (r.index[name] + 1) % len(instances)
	inst := instances[idx]

	return &inst, nil
}

func (r *Resolver) GetBreaker(key string) *gobreaker.CircuitBreaker {
	r.mu.Lock()
	defer r.mu.Unlock()

	if b, ok := r.breakers[key]; ok {
		return b
	}

	b := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        key,
		MaxRequests: 1,
		Timeout:     5 * time.Second,
		Interval:    10 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures >= 3
		},
	})
	r.breakers[key] = b
	return b
}
