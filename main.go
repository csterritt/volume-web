package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	cli "github.com/jawher/mow.cli"

	"github.com/gofiber/fiber/v2"

	"volume-web/weather"
)

const (
	volumeFilePath = "/Users/chris/tmp/volume.json"
	volumeStep     = 5
)

type VolumeState struct {
	Volume int  `json:"volume"`
	Muted  bool `json:"muted"`
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

var (
	mu     sync.Mutex
	state  VolumeState
)

const (
	defaultLat = 39.0438
	defaultLon = -77.4874
)

func main() {
	app := cli.App("volume-web", "Volume control server and weather tool")

	app.Command("serve", "Start the volume control web server", cmdServe)
	app.Command("weather", "Display current weather and forecast", cmdWeather)
	app.Command("weather-json", "Display current weather and forecast as JSON", cmdWeatherJSON)

	app.Run(os.Args)
}

func cmdServe(cmd *cli.Cmd) {
	cmd.Action = func() {
		startServer()
	}
}

func cmdWeather(cmd *cli.Cmd) {
	cmd.Action = func() {
		resp, err := weather.GetWeather(defaultLat, defaultLon)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching weather: %v\n", err)
			cli.Exit(1)
		}
		fmt.Print(weather.FormatWeather(resp))
	}
}

func cmdWeatherJSON(cmd *cli.Cmd) {
	cmd.Action = func() {
		resp, err := weather.GetWeather(defaultLat, defaultLon)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching weather: %v\n", err)
			cli.Exit(1)
		}
		output, err := weather.FormatWeatherJSON(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting weather JSON: %v\n", err)
			cli.Exit(1)
		}
		fmt.Println(output)
	}
}

func startServer() {
	if err := initializeVolumeState(); err != nil {
		fmt.Printf("Failed to initialize volume state: %v\n", err)
		return
	}

	app := fiber.New()

	app.Post("/api/v1/volume-up", handleVolumeUp)
	app.Post("/api/v1/volume-down", handleVolumeDown)
	app.Post("/api/v1/mute", handleMute)

	fmt.Println("Volume control server starting on :3400")
	if err := app.Listen(":3400"); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func initializeVolumeState() error {
	mu.Lock()
	defer mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(volumeFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Try to read existing state
	if data, err := os.ReadFile(volumeFilePath); err == nil {
		if err := json.Unmarshal(data, &state); err == nil {
			// Restore volume and mute state
			if err := setSystemVolume(state.Volume); err != nil {
				fmt.Printf("Warning: failed to restore volume: %v\n", err)
			}
			if err := setSystemMute(state.Muted); err != nil {
				fmt.Printf("Warning: failed to restore mute state: %v\n", err)
			}
			return nil
		}
	}

	// Get current system volume
	cmd := exec.Command("osascript", "-e", "set ovol to output volume of (get volume settings)")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current volume: %w", err)
	}

	volumeStr := string(output)
	volumeStr = volumeStr[:len(volumeStr)-1] // Remove newline
	volume, err := strconv.Atoi(volumeStr)
	if err != nil {
		return fmt.Errorf("failed to parse volume: %w", err)
	}

	state = VolumeState{
		Volume: volume,
		Muted:  false,
	}

	return saveVolumeState()
}

func handleVolumeUp(c *fiber.Ctx) error {
	mu.Lock()
	defer mu.Unlock()

	newVolume := state.Volume + volumeStep
	if newVolume > 100 {
		newVolume = 100
	}

	state.Volume = newVolume

	if err := saveVolumeState(); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	if err := setSystemVolume(newVolume); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	return c.JSON(Response{Success: true})
}

func handleVolumeDown(c *fiber.Ctx) error {
	mu.Lock()
	defer mu.Unlock()

	newVolume := state.Volume - volumeStep
	if newVolume < 0 {
		newVolume = 0
	}

	state.Volume = newVolume

	if err := saveVolumeState(); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	if err := setSystemVolume(newVolume); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	return c.JSON(Response{Success: true})
}

func handleMute(c *fiber.Ctx) error {
	mu.Lock()
	defer mu.Unlock()

	state.Muted = !state.Muted

	if err := saveVolumeState(); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	if err := setSystemMute(state.Muted); err != nil {
		return c.JSON(Response{Success: false, Error: err.Error()})
	}

	return c.JSON(Response{Success: true})
}

func saveVolumeState() error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(volumeFilePath, data, 0644)
}

func setSystemVolume(volume int) error {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf("set volume output volume %d", volume))
	return cmd.Run()
}

func setSystemMute(muted bool) error {
	muteValue := 0
	if muted {
		muteValue = 1
	}
	cmd := exec.Command("osascript", "-e", fmt.Sprintf("set volume output muted %d", muteValue))
	return cmd.Run()
}
