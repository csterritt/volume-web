package weather

import (
	"fmt"
	"sync"
	"time"
)

// FetchFunc is a function that retrieves weather data.
type FetchFunc func() (*WeatherResponse, error)

// WeatherCache holds cached weather data and refreshes it on a timer.
type WeatherCache struct {
	mu      sync.RWMutex
	data    *WeatherResponse
	fetchFn FetchFunc
	stopCh  chan struct{}
}

// NewWeatherCache creates a new cache that fetches weather data immediately
// and then refreshes every interval. Use Stop() to cancel the background goroutine.
func NewWeatherCache(fetchFn FetchFunc, interval time.Duration) *WeatherCache {
	c := &WeatherCache{
		fetchFn: fetchFn,
		stopCh:  make(chan struct{}),
	}

	go c.run(interval)
	return c
}

func (c *WeatherCache) run(interval time.Duration) {
	c.refresh()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.refresh()
		case <-c.stopCh:
			return
		}
	}
}

func (c *WeatherCache) refresh() {
	resp, err := c.fetchFn()
	if err != nil {
		fmt.Printf("Weather cache refresh failed: %v\n", err)
		return
	}

	c.mu.Lock()
	c.data = resp
	c.mu.Unlock()
}

// Get returns the cached weather data, or nil if not yet available.
func (c *WeatherCache) Get() *WeatherResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

// Stop cancels the background refresh goroutine.
func (c *WeatherCache) Stop() {
	select {
	case <-c.stopCh:
		// already stopped
	default:
		close(c.stopCh)
	}
}
