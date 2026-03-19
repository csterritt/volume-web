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

func TestFormatWeatherHasMultipleLines(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	output := FormatWeather(&resp)
	lines := strings.Split(output, "\n")

	// Should have many lines (header + current + forecast rows + spacing)
	assert.Greater(t, len(lines), 10)
}
