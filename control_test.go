/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClient reads one command and replies with the given result.
func fakeClient(t *testing.T, conn net.Conn, reply CommandResult, got chan<- Command) {
	t.Helper()
	line, err := bufio.NewReader(conn).ReadBytes('\n')
	require.NoError(t, err)
	var cmd Command
	require.NoError(t, json.Unmarshal(line, &cmd))
	got <- cmd
	payload, _ := json.Marshal(reply)
	_, err = conn.Write(append(payload, '\n'))
	require.NoError(t, err)
}

func TestControlServerNoClient(t *testing.T) {
	s := &ControlServer{}
	err := s.send(CmdMute)
	assert.ErrorIs(t, err, errNoClient)
}

func TestControlServerForwardsCommand(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	s := &ControlServer{}
	s.setClient(server)

	got := make(chan Command, 1)
	go fakeClient(t, client, CommandResult{Success: true}, got)

	require.NoError(t, s.send(CmdVolumeUp))
	select {
	case cmd := <-got:
		assert.Equal(t, CmdVolumeUp, cmd.Name)
	case <-time.After(time.Second):
		t.Fatal("client never received command")
	}
}

func TestControlServerPropagatesClientError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	s := &ControlServer{}
	s.setClient(server)

	got := make(chan Command, 1)
	go fakeClient(t, client, CommandResult{Success: false, Error: "osascript failed"}, got)

	err := s.send(CmdVolumeDown)
	require.Error(t, err)
	assert.Equal(t, "osascript failed", err.Error())
	assert.Equal(t, CmdVolumeDown, (<-got).Name)
}

func TestControlServerDropsClientOnDisconnect(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()

	s := &ControlServer{}
	s.setClient(server)

	client.Close()

	err := s.send(CmdMute)
	require.Error(t, err)
	assert.NotErrorIs(t, err, errNoClient)

	// Client should now be gone
	assert.ErrorIs(t, s.send(CmdMute), errNoClient)
}

func TestClientServeConnectionUnknownCommand(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	done := make(chan error, 1)
	go func() { done <- serveConnection(client) }()

	payload, _ := json.Marshal(Command{Name: "bogus"})
	_, err := server.Write(append(payload, '\n'))
	require.NoError(t, err)

	line, err := bufio.NewReader(server).ReadBytes('\n')
	require.NoError(t, err)
	var result CommandResult
	require.NoError(t, json.Unmarshal(line, &result))
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unknown command")

	server.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("serveConnection did not exit on close")
	}
}
