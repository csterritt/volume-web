package weather

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatWeatherContainsCurrentConditions(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output := FormatWeather(&resp)

	// Should contain current weather header
	assert.Contains(t, output, "Current Weather")

	// Should contain temperature
	assert.Contains(t, output, "0.7")

	// Should contain apparent temperature
	assert.Contains(t, output, "-3.6")

	// Should contain humidity
	assert.Contains(t, output, "59")

	// Should contain wind speed
	assert.Contains(t, output, "8.9")
}

func TestFormatWeatherContainsForecast(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output := FormatWeather(&resp)

	// Should contain forecast header
	assert.Contains(t, output, "7-Day Forecast")

	// Should contain all 7 forecast dates
	assert.Contains(t, output, "2026-03-18")
	assert.Contains(t, output, "2026-03-24")

	// Should contain high/low temps
	assert.Contains(t, output, "4.6")
	assert.Contains(t, output, "-4.4")

	// Should contain precipitation probability
	assert.Contains(t, output, "4%")
}

func TestFormatWeatherContainsLocation(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output := FormatWeather(&resp)

	// Should contain timezone
	assert.Contains(t, output, "America/New_York")

	// Should contain elevation
	assert.Contains(t, output, "93")
}

func TestFormatWeatherJSONIsValidJSON(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output, err := FormatWeatherJSON(&resp)
	require.NoError(t, err)

	// Should be valid JSON that round-trips back to a WeatherResponse
	var roundTrip WeatherResponse
	err = json.Unmarshal([]byte(output), &roundTrip)
	require.NoError(t, err)

	assert.InDelta(t, resp.Latitude, roundTrip.Latitude, 0.0001)
	assert.Equal(t, resp.Timezone, roundTrip.Timezone)
	assert.InDelta(t, resp.Current.Temperature, roundTrip.Current.Temperature, 0.01)
	assert.Equal(t, resp.Current.Humidity, roundTrip.Current.Humidity)
	assert.Len(t, roundTrip.Daily.Time, 7)
	assert.Equal(t, resp.Daily.TemperatureMax, roundTrip.Daily.TemperatureMax)
}

func TestFormatWeatherJSONIsIndented(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output, err := FormatWeatherJSON(&resp)
	require.NoError(t, err)

	// Indented JSON should contain newlines and leading spaces
	assert.Contains(t, output, "\n")
	assert.Contains(t, output, "  ")
}

func TestFormatWeatherHasMultipleLines(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output := FormatWeather(&resp)
	lines := strings.Split(output, "\n")

	// Should have many lines (header + current + forecast rows + spacing)
	assert.Greater(t, len(lines), 10)
}
