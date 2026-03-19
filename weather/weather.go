package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.open-meteo.com/v1/forecast"
	currentParams  = "temperature_2m,apparent_temperature,relative_humidity_2m,surface_pressure,wind_speed_10m,wind_direction_10m,weather_code"
	dailyParams    = "temperature_2m_max,temperature_2m_min,precipitation_probability_max,weather_code"
)

// GetWeather retrieves current and forecast weather data for the given coordinates.
func GetWeather(lat, lon float64) (*WeatherResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return GetWeatherWithClient(client, defaultBaseURL, lat, lon)
}

// GetWeatherWithClient retrieves weather data using a custom HTTP client and base URL.
// This is primarily used for testing with mock servers.
func GetWeatherWithClient(client *http.Client, baseURL string, lat, lon float64) (*WeatherResponse, error) {
	url := fmt.Sprintf("%s?latitude=%f&longitude=%f&current=%s&daily=%s&timezone=auto",
		baseURL, lat, lon, currentParams, dailyParams)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %w", err)
	}

	return &weather, nil
}
