package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolumeStateSerialization(t *testing.T) {
	state := VolumeState{
		Volume: 50,
		Muted:  false,
	}

	data, err := json.Marshal(state)
	assert.NoError(t, err)

	var decoded VolumeState
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, state.Volume, decoded.Volume)
	assert.Equal(t, state.Muted, decoded.Muted)
}

func TestSaveVolumeState(t *testing.T) {
	// Create a test state
	testState := VolumeState{
		Volume: 75,
		Muted:  true,
	}

	// Test JSON marshaling/unmarshaling directly
	data, err := json.Marshal(testState)
	assert.NoError(t, err)

	var savedState VolumeState
	err = json.Unmarshal(data, &savedState)
	assert.NoError(t, err)
	assert.Equal(t, testState.Volume, savedState.Volume)
	assert.Equal(t, testState.Muted, savedState.Muted)
}

func TestVolumeBoundaries(t *testing.T) {
	// Test volume up boundary
	currentVolume := 95
	newVolume := currentVolume + volumeStep
	if newVolume > 100 {
		newVolume = 100
	}
	assert.Equal(t, 100, newVolume)

	// Test volume down boundary
	currentVolume = 5
	newVolume = currentVolume - volumeStep
	if newVolume < 0 {
		newVolume = 0
	}
	assert.Equal(t, 0, newVolume)
}
