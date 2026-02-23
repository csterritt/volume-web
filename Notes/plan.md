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
