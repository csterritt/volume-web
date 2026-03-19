# Volume Control Web Server Implementation Plan

## Overview
Create a Go web server using Fiber framework to control system volume via HTTP endpoints.

## Implementation Steps

### 1. Project Setup
- Add Fiber dependency to go.mod
- Create volume state struct for JSON marshaling
- Set up basic Fiber server structure

### 2. Volume State Management
- Create VolumeState struct with Volume (int) and Muted (bool) fields
- Implement file I/O operations for `/Users/chris/tmp/volume.json`
- Add shell command execution functions for osascript calls

### 3. Server Initialization
- On startup: check if volume.json exists
- If not exists: run `osascript -e 'set ovol to output volume of (get volume settings)'` to get current volume
- Initialize volume state file with current volume and muted=false
- If exists: read file and restore volume/mute state via shell commands

### 4. API Endpoints
- POST `/api/v1/volume-up`: Increase volume by 10%, cap at 100
- POST `/api/v1/volume-down`: Decrease volume by 10%, floor at 0  
- POST `/api/v1/mute`: Toggle mute state

### 5. Response Handling
- Success response: `{"success": true}`
- Error response: `{"success": false, "error": "message"}`

### 6. Volume Control Logic
For each endpoint:
1. Read current state from file
2. Update volume/mute values
3. Write updated state back to file
4. Execute appropriate osascript command:
   - Volume: `osascript -e "set volume output volume {level}"`
   - Mute: `osascript -e "set volume output muted {0|1}"`

## Pitfalls & Considerations
- File permission issues with `/Users/chris/tmp/` directory
- Shell command execution errors handling
- Concurrent request handling (file locking)
- Volume range validation (0-100)
- Error handling for osascript failures

## Testing Strategy
- Unit tests for file I/O operations
- Integration tests for API endpoints
- Mock shell commands for testing
- Test edge cases (volume boundaries, file missing scenarios)

---

# Weather Package Implementation Plan

## Overview
Add a `weather` sub-package that retrieves current conditions and a 7-day daily forecast from the Open-Meteo API.

## API Details
- Endpoint: `https://api.open-meteo.com/v1/forecast`
- Current params: `temperature_2m`, `apparent_temperature`, `relative_humidity_2m`, `surface_pressure`, `wind_speed_10m`, `wind_direction_10m`, `weather_code`
- Daily params: `temperature_2m_max`, `temperature_2m_min`, `precipitation_probability_max`, `weather_code`
- Timezone: `auto`
- No API key required

## Package Structure
```
weather/
├── types.go          # Data structs for API response
├── weather.go        # GetWeather function + HTTP client logic
└── weather_test.go   # Unit tests with mock HTTP responses
```

## Implementation Steps

### 1. Define Types (`types.go`)
- `CurrentWeather`: temperature, apparent_temperature, humidity, pressure, wind speed/direction, weather code
- `DailyForecast`: arrays of time, max/min temp, precipitation probability, weather code
- `WeatherResponse`: wraps current + daily + metadata (lat, lon, timezone, elevation)
- Internal API response structs matching Open-Meteo JSON shape

### 2. Write Failing Tests (`weather_test.go`) — Red Phase
- Test JSON parsing of a known API response into structs
- Test `GetWeather` with a mock HTTP server returning canned JSON
- Test error handling: bad JSON, HTTP errors, invalid coordinates

### 3. Implement `GetWeather` (`weather.go`) — Green Phase
- `GetWeather(lat, lon float64) (*WeatherResponse, error)`
- Build URL with query parameters
- Use `net/http` with a 10-second timeout
- Parse JSON into structs, return `*WeatherResponse` or error
- Accept an optional `HTTPClient` interface for testability

### 4. Run Tests & Verify

## Pitfalls
- Open-Meteo returns `float64` for most numeric fields but some (humidity, weather_code) are `int`
- API may return slightly adjusted lat/lon vs. what was requested
- Network timeouts need explicit handling
- No API key, but rate limits may apply for heavy usage
