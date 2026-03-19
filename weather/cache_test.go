package weather

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSampleResponse() *WeatherResponse {
	var resp WeatherResponse
	json.Unmarshal([]byte(sampleAPIResponse), &resp)
	return &resp
}

func TestWeatherCacheGetReturnsNilBeforeFetch(t *testing.T) {
	// A cache with a fetch function that blocks forever should return nil immediately
	blockCh := make(chan struct{})
	fetchFn := func() (*WeatherResponse, error) {
		<-blockCh
		return makeSampleResponse(), nil
	}

	cache := NewWeatherCache(fetchFn, 1*time.Hour)
	defer cache.Stop()

	// Give a tiny moment but not enough for the blocking fetch
	time.Sleep(10 * time.Millisecond)

	// Cache should still be nil since fetch hasn't completed
	// (it's blocked)
	result := cache.Get()
	assert.Nil(t, result)

	close(blockCh)
}

func TestWeatherCacheGetReturnsCachedData(t *testing.T) {
	sample := makeSampleResponse()
	fetchFn := func() (*WeatherResponse, error) {
		return sample, nil
	}

	cache := NewWeatherCache(fetchFn, 1*time.Hour)
	defer cache.Stop()

	// Wait for initial fetch
	assert.Eventually(t, func() bool {
		return cache.Get() != nil
	}, 2*time.Second, 10*time.Millisecond)

	result := cache.Get()
	require.NotNil(t, result)
	assert.InDelta(t, sample.Latitude, result.Latitude, 0.0001)
	assert.Equal(t, sample.Timezone, result.Timezone)
	assert.InDelta(t, sample.Current.Temperature, result.Current.Temperature, 0.01)
}

func TestWeatherCacheRefreshesOnInterval(t *testing.T) {
	var callCount atomic.Int32
	fetchFn := func() (*WeatherResponse, error) {
		callCount.Add(1)
		return makeSampleResponse(), nil
	}

	cache := NewWeatherCache(fetchFn, 50*time.Millisecond)
	defer cache.Stop()

	// Wait for at least 3 fetches (initial + 2 refreshes)
	assert.Eventually(t, func() bool {
		return callCount.Load() >= 3
	}, 2*time.Second, 10*time.Millisecond)
}

func TestWeatherCacheHandlesFetchError(t *testing.T) {
	callCount := atomic.Int32{}
	fetchFn := func() (*WeatherResponse, error) {
		count := callCount.Add(1)
		if count == 1 {
			return makeSampleResponse(), nil
		}
		return nil, assert.AnError
	}

	cache := NewWeatherCache(fetchFn, 50*time.Millisecond)
	defer cache.Stop()

	// Wait for initial fetch to succeed
	assert.Eventually(t, func() bool {
		return cache.Get() != nil
	}, 2*time.Second, 10*time.Millisecond)

	// Wait for a failed refresh
	time.Sleep(150 * time.Millisecond)

	// Cache should still have original data (not cleared on error)
	result := cache.Get()
	require.NotNil(t, result)
	assert.InDelta(t, makeSampleResponse().Latitude, result.Latitude, 0.0001)
}

func TestWeatherCacheStopPreventsMoreFetches(t *testing.T) {
	var callCount atomic.Int32
	fetchFn := func() (*WeatherResponse, error) {
		callCount.Add(1)
		return makeSampleResponse(), nil
	}

	cache := NewWeatherCache(fetchFn, 50*time.Millisecond)

	// Wait for initial fetch
	assert.Eventually(t, func() bool {
		return cache.Get() != nil
	}, 2*time.Second, 10*time.Millisecond)

	cache.Stop()
	countAfterStop := callCount.Load()

	// Wait a bit and ensure no more fetches
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, countAfterStop, callCount.Load())
}
