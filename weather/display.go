package weather

import (
	"fmt"
	"strings"
)

// FormatWeather formats weather data for terminal display.
func FormatWeather(w *WeatherResponse) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Location: %.4f, %.4f | Timezone: %s | Elevation: %.0fm\n",
		w.Latitude, w.Longitude, w.Timezone, w.Elevation)
	b.WriteString("\n")

	// Current conditions
	b.WriteString("=== Current Weather ===\n")
	fmt.Fprintf(&b, "  Time:          %s\n", w.Current.Time)
	fmt.Fprintf(&b, "  Temperature:   %.1f%s\n", w.Current.Temperature, w.CurrentUnits.Temperature)
	fmt.Fprintf(&b, "  Feels Like:    %.1f%s\n", w.Current.ApparentTemperature, w.CurrentUnits.ApparentTemperature)
	fmt.Fprintf(&b, "  Humidity:      %d%s\n", w.Current.Humidity, w.CurrentUnits.Humidity)
	fmt.Fprintf(&b, "  Wind:          %.1f %s from %d%s\n",
		w.Current.WindSpeed, w.CurrentUnits.WindSpeed,
		w.Current.WindDirection, w.CurrentUnits.WindDirection)
	fmt.Fprintf(&b, "  Condition:     %s\n", describeWeatherCode(w.Current.WeatherCode))
	b.WriteString("\n")

	// Daily forecast
	b.WriteString("=== 7-Day Forecast ===\n")
	fmt.Fprintf(&b, "  %-12s  %6s  %6s  %6s  %s\n", "Date", "High", "Low", "Precip", "Condition")
	fmt.Fprintf(&b, "  %-12s  %6s  %6s  %6s  %s\n", "----", "----", "---", "------", "---------")

	days := len(w.Daily.Time)
	for i := 0; i < days; i++ {
		fmt.Fprintf(&b, "  %-12s  %5.1f%s  %5.1f%s  %4d%%   %s\n",
			w.Daily.Time[i],
			w.Daily.TemperatureMax[i], w.DailyUnits.TemperatureMax,
			w.Daily.TemperatureMin[i], w.DailyUnits.TemperatureMin,
			w.Daily.PrecipitationProbability[i],
			describeWeatherCode(w.Daily.WeatherCode[i]),
		)
	}

	return b.String()
}

// describeWeatherCode returns a human-readable description for a WMO weather code.
func describeWeatherCode(code int) string {
	switch code {
	case 0:
		return "Clear sky"
	case 1:
		return "Mainly clear"
	case 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45, 48:
		return "Fog"
	case 51:
		return "Light drizzle"
	case 53:
		return "Moderate drizzle"
	case 55:
		return "Dense drizzle"
	case 61:
		return "Slight rain"
	case 63:
		return "Moderate rain"
	case 65:
		return "Heavy rain"
	case 71:
		return "Slight snow"
	case 73:
		return "Moderate snow"
	case 75:
		return "Heavy snow"
	case 77:
		return "Snow grains"
	case 80:
		return "Slight rain showers"
	case 81:
		return "Moderate rain showers"
	case 82:
		return "Violent rain showers"
	case 85:
		return "Slight snow showers"
	case 86:
		return "Heavy snow showers"
	case 95:
		return "Thunderstorm"
	case 96, 99:
		return "Thunderstorm with hail"
	default:
		return fmt.Sprintf("Unknown (%d)", code)
	}
}
