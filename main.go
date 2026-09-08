/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	cli "github.com/jawher/mow.cli"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"volume-web/weather"
)

const (
	volumeFilePath = "/Users/chris/tmp/volume.json"
	volumeStep     = 3

	defaultHTTPPort    = 3400
	defaultControlPort = 3401

	commandTimeout   = 5 * time.Second
	reconnectMinWait = 1 * time.Second
	reconnectMaxWait = 30 * time.Second
)

// Volume command names sent from server to client.
const (
	CmdVolumeUp   = "volume-up"
	CmdVolumeDown = "volume-down"
	CmdMute       = "mute"
)

type VolumeState struct {
	Volume int  `json:"volume"`
	Muted  bool `json:"muted"`
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Command is sent from server to client over the control connection.
type Command struct {
	Name string `json:"name"`
}

// CommandResult is sent from client back to server over the control connection.
type CommandResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

var (
	mu    sync.Mutex
	state VolumeState
)

const (
	defaultLat = 51.5085
	defaultLon = -0.1257
)

func main() {
	app := cli.App("volume-web", "Volume control server and weather tool")

	app.Command("server", "Serve weather data and accept volume commands, forwarding them to a connected client", cmdServer)
	app.Command("client", "Connect to a server and execute forwarded volume commands", cmdClient)

	app.Run(os.Args)
}

func cmdServer(cmd *cli.Cmd) {
	cmd.Spec = "[--port] [--control-port]"

	httpPort := cmd.IntOpt("p port", defaultHTTPPort, "HTTP port for weather and volume API")
	controlPort := cmd.IntOpt("c control-port", defaultControlPort, "TCP port for client control connections")

	cmd.Action = func() {
		startServer(*httpPort, *controlPort)
	}
}

func cmdClient(cmd *cli.Cmd) {
	cmd.Spec = "[--server]"

	serverAddr := cmd.StringOpt("s server", fmt.Sprintf("localhost:%d", defaultControlPort), "Server control address (host:port)")

	cmd.Action = func() {
		startClient(*serverAddr)
	}
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

// ControlServer accepts a client connection and forwards volume commands to it.
type ControlServer struct {
	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
}

var errNoClient = errors.New("no volume client connected")

func (s *ControlServer) listen(port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Control listener failed to start: %v\n", err)
		return
	}
	fmt.Printf("Control listener started on %s\n", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Control accept error: %v\n", err)
			continue
		}
		s.setClient(conn)
	}
}

func (s *ControlServer) setClient(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		fmt.Printf("Replacing existing client %s with %s\n", s.conn.RemoteAddr(), conn.RemoteAddr())
		s.conn.Close()
	} else {
		fmt.Printf("Client connected from %s\n", conn.RemoteAddr())
	}
	s.conn = conn
	s.reader = bufio.NewReader(conn)
}

func (s *ControlServer) dropClient() {
	if s.conn != nil {
		fmt.Printf("Client %s disconnected\n", s.conn.RemoteAddr())
		s.conn.Close()
	}
	s.conn = nil
	s.reader = nil
}

// send forwards a command to the connected client and waits for its result.
func (s *ControlServer) send(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return errNoClient
	}

	payload, err := json.Marshal(Command{Name: name})
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}
	payload = append(payload, '\n')

	deadline := time.Now().Add(commandTimeout)
	s.conn.SetDeadline(deadline)
	defer s.conn.SetDeadline(time.Time{})

	if _, err := s.conn.Write(payload); err != nil {
		s.dropClient()
		return fmt.Errorf("failed to send command to client: %w", err)
	}

	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		s.dropClient()
		return fmt.Errorf("failed to read result from client: %w", err)
	}

	var result CommandResult
	if err := json.Unmarshal(line, &result); err != nil {
		return fmt.Errorf("invalid result from client: %w", err)
	}
	if !result.Success {
		return errors.New(result.Error)
	}
	return nil
}

func startServer(httpPort, controlPort int) {
	wCache := weather.NewWeatherCache(func() (*weather.WeatherResponse, error) {
		return weather.GetWeather(defaultLat, defaultLon)
	}, 10*time.Minute)
	defer wCache.Stop()

	control := &ControlServer{}
	go control.listen(controlPort)

	app := fiber.New()
	app.Use(logger.New())

	app.Post("/api/v1/volume-up", forwardHandler(control, CmdVolumeUp))
	app.Post("/api/v1/volume-down", forwardHandler(control, CmdVolumeDown))
	app.Post("/api/v1/mute", forwardHandler(control, CmdMute))
	app.Get("/weather", handleWeather(wCache))

	addr := fmt.Sprintf(":%d", httpPort)
	fmt.Printf("Volume control server starting on %s\n", addr)
	if err := app.Listen(addr); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func forwardHandler(control *ControlServer, name string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := control.send(name); err != nil {
			status := fiber.StatusBadGateway
			if errors.Is(err, errNoClient) {
				status = fiber.StatusServiceUnavailable
			}
			return c.Status(status).JSON(Response{Success: false, Error: err.Error()})
		}
		return c.JSON(Response{Success: true})
	}
}

type WeatherEndpointResponse struct {
	Weather   *weather.WeatherResponse `json:"weather"`
	Timestamp string                   `json:"timestamp"`
}

func handleWeather(cache *weather.WeatherCache) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := cache.Get()
		if data == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(Response{
				Success: false,
				Error:   "weather data not yet available",
			})
		}
		return c.JSON(WeatherEndpointResponse{
			Weather:   data,
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

func startClient(serverAddr string) {
	if err := initializeVolumeState(); err != nil {
		fmt.Printf("Failed to initialize volume state: %v\n", err)
		return
	}

	wait := reconnectMinWait
	for {
		fmt.Printf("Connecting to server at %s\n", serverAddr)
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			fmt.Printf("Connection failed: %v (retrying in %s)\n", err, wait)
			time.Sleep(wait)
			wait *= 2
			if wait > reconnectMaxWait {
				wait = reconnectMaxWait
			}
			continue
		}

		wait = reconnectMinWait
		fmt.Printf("Connected to server at %s\n", serverAddr)
		if err := serveConnection(conn); err != nil {
			fmt.Printf("Connection lost: %v\n", err)
		}
		conn.Close()
		time.Sleep(wait)
	}
}

// serveConnection reads commands from the server, executes them, and replies
// with results until the connection fails.
func serveConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}

		var cmd Command
		result := CommandResult{Success: true}
		if err := json.Unmarshal(line, &cmd); err != nil {
			result = CommandResult{Success: false, Error: fmt.Sprintf("invalid command: %v", err)}
		} else if err := executeCommand(cmd.Name); err != nil {
			result = CommandResult{Success: false, Error: err.Error()}
		}

		fmt.Printf("Command %q -> success=%v %s\n", cmd.Name, result.Success, result.Error)

		payload, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		payload = append(payload, '\n')
		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}
}

func executeCommand(name string) error {
	switch name {
	case CmdVolumeUp:
		return volumeUp()
	case CmdVolumeDown:
		return volumeDown()
	case CmdMute:
		return toggleMute()
	default:
		return fmt.Errorf("unknown command: %q", name)
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

	volume, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse volume: %w", err)
	}

	state = VolumeState{
		Volume: volume,
		Muted:  false,
	}

	return saveVolumeState()
}

func volumeUp() error {
	mu.Lock()
	defer mu.Unlock()

	newVolume := state.Volume + volumeStep
	if newVolume > 100 {
		newVolume = 100
	}

	state.Volume = newVolume

	if err := saveVolumeState(); err != nil {
		return err
	}
	return setSystemVolume(newVolume)
}

func volumeDown() error {
	mu.Lock()
	defer mu.Unlock()

	newVolume := state.Volume - volumeStep
	if newVolume < 0 {
		newVolume = 0
	}

	state.Volume = newVolume

	if err := saveVolumeState(); err != nil {
		return err
	}
	return setSystemVolume(newVolume)
}

func toggleMute() error {
	mu.Lock()
	defer mu.Unlock()

	state.Muted = !state.Muted

	if err := saveVolumeState(); err != nil {
		return err
	}
	return setSystemMute(state.Muted)
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
