package observability

import (
	"math"
	"sort"
	"sync"
	"time"
)

const latencyWindowSize = 512

type routeMetrics struct {
	count   int64
	errors  int64
	latency []float64
	next    int
	filled  bool
}

type Collector struct {
	mu             sync.Mutex
	routes         map[string]*routeMetrics
	lockContention int64
	queueLag       int64
}

type RouteSnapshot struct {
	Route string  `json:"route"`
	Count int64   `json:"count"`
	Errors int64  `json:"errors"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

type Snapshot struct {
	GeneratedAtUTC string          `json:"generated_at_utc"`
	TotalRequests  int64           `json:"total_requests"`
	TotalErrors    int64           `json:"total_errors"`
	ErrorRate      float64         `json:"error_rate"`
	QueueLag       int64           `json:"queue_lag"`
	LockContention int64           `json:"lock_contention"`
	Routes         []RouteSnapshot `json:"routes"`
}

var defaultCollector = &Collector{
	routes: make(map[string]*routeMetrics),
}

func ObserveRequest(route string, status int, duration time.Duration) {
	defaultCollector.ObserveRequest(route, status, duration)
}

func IncLockContention() {
	defaultCollector.IncLockContention()
}

func SetQueueLag(value int64) {
	defaultCollector.SetQueueLag(value)
}

func GetSnapshot() Snapshot {
	return defaultCollector.Snapshot()
}

func (c *Collector) ObserveRequest(route string, status int, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := c.routes[route]
	if item == nil {
		item = &routeMetrics{latency: make([]float64, latencyWindowSize)}
		c.routes[route] = item
	}

	item.count++
	if status >= 500 {
		item.errors++
	}

	item.latency[item.next] = duration.Seconds() * 1000
	item.next = (item.next + 1) % len(item.latency)
	if item.next == 0 {
		item.filled = true
	}
}

func (c *Collector) IncLockContention() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lockContention++
}

func (c *Collector) SetQueueLag(value int64) {
	if value < 0 {
		value = 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queueLag = value
}

func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	routes := make([]RouteSnapshot, 0, len(c.routes))
	var totalRequests int64
	var totalErrors int64

	for route, metric := range c.routes {
		sample := copyLatencySamples(metric)
		p95 := percentile(sample, 95)
		p99 := percentile(sample, 99)

		routes = append(routes, RouteSnapshot{
			Route:  route,
			Count:  metric.count,
			Errors: metric.errors,
			P95Ms:  round2(p95),
			P99Ms:  round2(p99),
		})

		totalRequests += metric.count
		totalErrors += metric.errors
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Route < routes[j].Route
	})

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = (float64(totalErrors) / float64(totalRequests)) * 100
	}

	return Snapshot{
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
		TotalRequests:  totalRequests,
		TotalErrors:    totalErrors,
		ErrorRate:      round2(errorRate),
		QueueLag:       c.queueLag,
		LockContention: c.lockContention,
		Routes:         routes,
	}
}

func copyLatencySamples(metric *routeMetrics) []float64 {
	size := metric.next
	if metric.filled {
		size = len(metric.latency)
	}
	if size == 0 {
		return nil
	}

	result := make([]float64, 0, size)
	if metric.filled {
		result = append(result, metric.latency[metric.next:]...)
		result = append(result, metric.latency[:metric.next]...)
		return result
	}
	result = append(result, metric.latency[:metric.next]...)
	return result
}

func percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}
	copied := append([]float64(nil), values...)
	sort.Float64s(copied)

	rank := int(math.Ceil((float64(p)/100.0)*float64(len(copied)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(copied) {
		rank = len(copied) - 1
	}
	return copied[rank]
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
