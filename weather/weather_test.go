package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleAPIResponse = `{
	"latitude": 39.051937,
	"longitude": -77.47853,
	"generationtime_ms": 0.455,
	"utc_offset_seconds": -14400,
	"timezone": "America/New_York",
	"timezone_abbreviation": "GMT-4",
	"elevation": 93.0,
	"current_units": {
		"time": "iso8601",
		"interval": "seconds",
		"temperature_2m": "°C",
		"apparent_temperature": "°C",
		"relative_humidity_2m": "%",
		"wind_speed_10m": "km/h",
		"wind_direction_10m": "°",
		"weather_code": "wmo code"
	},
	"current": {
		"time": "2026-03-18T21:15",
		"temperature_2m": 0.7,
		"apparent_temperature": -3.6,
		"relative_humidity_2m": 59,
		"wind_speed_10m": 8.9,
		"wind_direction_10m": 122,
		"weather_code": 0
	},
	"daily_units": {
		"time": "iso8601",
		"temperature_2m_max": "°C",
		"temperature_2m_min": "°C",
		"precipitation_probability_max": "%",
		"weather_code": "wmo code"
	},
	"daily": {
		"time": ["2026-03-18", "2026-03-19", "2026-03-20", "2026-03-21", "2026-03-22", "2026-03-23", "2026-03-24"],
		"temperature_2m_max": [4.6, 12.7, 16.9, 17.0, 21.5, 15.9, 8.1],
		"temperature_2m_min": [-4.4, -0.4, 0.3, 9.6, 7.8, 1.6, -2.7],
		"precipitation_probability_max": [4, 4, 36, 36, 30, 32, 8],
		"weather_code": [3, 3, 53, 63, 3, 53, 3]
	}
}`

func TestParseWeatherResponse(t *testing.T) {
	var resp WeatherResponse
	err := json.Unmarshal([]byte(sampleAPIResponse), &resp)
	require.NoError(t, err)

	assert.InDelta(t, 39.051937, resp.Latitude, 0.0001)
	assert.InDelta(t, -77.47853, resp.Longitude, 0.0001)
	assert.InDelta(t, 93.0, resp.Elevation, 0.1)
	assert.Equal(t, "America/New_York", resp.Timezone)
	assert.Equal(t, -14400, resp.UTCOffsetSeconds)

	// Current weather
	assert.Equal(t, "2026-03-18T21:15", resp.Current.Time)
	assert.InDelta(t, 0.7, resp.Current.Temperature, 0.01)
	assert.InDelta(t, -3.6, resp.Current.ApparentTemperature, 0.01)
	assert.Equal(t, 59, resp.Current.Humidity)
	assert.InDelta(t, 8.9, resp.Current.WindSpeed, 0.01)
	assert.Equal(t, 122, resp.Current.WindDirection)
	assert.Equal(t, 0, resp.Current.WeatherCode)

	// Current units
	assert.Equal(t, "°C", resp.CurrentUnits.Temperature)
	assert.Equal(t, "km/h", resp.CurrentUnits.WindSpeed)

	// Daily forecast
	assert.Len(t, resp.Daily.Time, 7)
	assert.Equal(t, "2026-03-18", resp.Daily.Time[0])
	assert.InDelta(t, 4.6, resp.Daily.TemperatureMax[0], 0.01)
	assert.InDelta(t, -4.4, resp.Daily.TemperatureMin[0], 0.01)
	assert.Equal(t, 4, resp.Daily.PrecipitationProbability[0])
	assert.Equal(t, 3, resp.Daily.WeatherCode[0])
}

func TestGetWeatherWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		q := r.URL.Query()
		assert.Equal(t, "39.043800", q.Get("latitude"))
		assert.Equal(t, "-77.487400", q.Get("longitude"))
		assert.Equal(t, "America/New_York", q.Get("timezone"))
		assert.NotEmpty(t, q.Get("current"))
		assert.NotEmpty(t, q.Get("daily"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleAPIResponse))
	}))
	defer server.Close()

	resp, err := GetWeatherWithClient(server.Client(), server.URL, 39.0438, -77.4874)
	require.NoError(t, err)

	assert.InDelta(t, 39.051937, resp.Latitude, 0.0001)
	assert.InDelta(t, 0.7, resp.Current.Temperature, 0.01)
	assert.Len(t, resp.Daily.Time, 7)
}

func TestGetWeatherHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": true, "reason": "internal error"}`))
	}))
	defer server.Close()

	_, err := GetWeatherWithClient(server.Client(), server.URL, 39.0438, -77.4874)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestGetWeatherBadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	_, err := GetWeatherWithClient(server.Client(), server.URL, 39.0438, -77.4874)
	assert.Error(t, err)
}

func TestGetWeatherServerUnreachable(t *testing.T) {
	_, err := GetWeatherWithClient(http.DefaultClient, "http://localhost:1", 39.0438, -77.4874)
	assert.Error(t, err)
}
