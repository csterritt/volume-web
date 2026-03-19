package weather

// CurrentWeather holds current weather conditions.
type CurrentWeather struct {
	Time               string  `json:"time"`
	Temperature        float64 `json:"temperature_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	Humidity           int     `json:"relative_humidity_2m"`
	Pressure           float64 `json:"surface_pressure"`
	WindSpeed          float64 `json:"wind_speed_10m"`
	WindDirection      int     `json:"wind_direction_10m"`
	WeatherCode        int     `json:"weather_code"`
}

// DailyForecast holds 7-day daily forecast data.
type DailyForecast struct {
	Time                     []string  `json:"time"`
	TemperatureMax           []float64 `json:"temperature_2m_max"`
	TemperatureMin           []float64 `json:"temperature_2m_min"`
	PrecipitationProbability []int     `json:"precipitation_probability_max"`
	WeatherCode              []int     `json:"weather_code"`
}

// CurrentUnits holds the units for current weather fields.
type CurrentUnits struct {
	Time               string `json:"time"`
	Interval           string `json:"interval"`
	Temperature        string `json:"temperature_2m"`
	ApparentTemperature string `json:"apparent_temperature"`
	Humidity           string `json:"relative_humidity_2m"`
	Pressure           string `json:"surface_pressure"`
	WindSpeed          string `json:"wind_speed_10m"`
	WindDirection      string `json:"wind_direction_10m"`
	WeatherCode        string `json:"weather_code"`
}

// DailyUnits holds the units for daily forecast fields.
type DailyUnits struct {
	Time                     string `json:"time"`
	TemperatureMax           string `json:"temperature_2m_max"`
	TemperatureMin           string `json:"temperature_2m_min"`
	PrecipitationProbability string `json:"precipitation_probability_max"`
	WeatherCode              string `json:"weather_code"`
}

// WeatherResponse is the top-level response from the Open-Meteo API.
type WeatherResponse struct {
	Latitude             float64        `json:"latitude"`
	Longitude            float64        `json:"longitude"`
	Elevation            float64        `json:"elevation"`
	GenerationTimeMs     float64        `json:"generationtime_ms"`
	UTCOffsetSeconds     int            `json:"utc_offset_seconds"`
	Timezone             string         `json:"timezone"`
	TimezoneAbbreviation string         `json:"timezone_abbreviation"`
	CurrentUnits         CurrentUnits   `json:"current_units"`
	Current              CurrentWeather `json:"current"`
	DailyUnits           DailyUnits     `json:"daily_units"`
	Daily                DailyForecast  `json:"daily"`
}
