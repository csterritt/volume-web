# Weather Package Implementation Plan

## Overview
Create a new Go package named `weather` that retrieves current and forecast weather data using the Open-Meteo API.

## Requirements

### Location
- Default coordinates: 39.0438° N, -77.4874° W (Ashburn, Virginia area)
- Package should accept any latitude/longitude parameters

### API Integration
- Use Open-Meteo API: https://open-meteo.com/en/docs
- Retrieve both current weather and forecast data
- Handle API errors and edge cases gracefully

### Package Structure
```
weather/
├── go.mod
├── weather.go
└── types.go
```

### Core Components

#### 1. Data Structures (`types.go`)
- `CurrentWeather` struct for current conditions
- `ForecastWeather` struct for forecast data
- `WeatherResponse` struct combining both
- Include relevant fields: temperature, humidity, wind, precipitation, etc.

#### 2. Main Function (`weather.go`)
- `GetWeather(lat, lon float64) (*WeatherResponse, error)`
- Construct API request with proper parameters
- Parse JSON response into Go structs
- Return structured data or error

#### 3. Error Handling
- Network timeout handling
- API rate limiting
- Invalid coordinates
- JSON parsing errors

### API Parameters to Include
- Current weather: temperature, apparent_temperature, humidity, pressure, wind speed/direction
- Daily forecast: max/min temperature, precipitation probability
- Hourly forecast (optional): temperature, precipitation
- Weather codes for conditions

### Testing
- Unit tests for data parsing
- Mock API responses for testing
- Error scenario testing

### Usage Example
```go
weather, err := GetWeather(39.0438, -77.4874)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current temperature: %.1f°C\n", weather.Current.Temperature)
```
