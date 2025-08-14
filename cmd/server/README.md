# Worldographer Map Server

A web server that displays information about Worldographer files (WXX files) including map metadata and hex grid visualization.

## Overview

This server loads a Worldographer file and provides a web interface to view:
- Map metadata (application version, data version, creation date)
- Map dimensions and hex configuration
- Interactive hex grid visualization (first 5x5 tiles)
- Coordinate information for each hex (row/column and cube coordinates)

## Building

```bash
go build -o dist/local/server ./cmd/server
```

## Usage

```bash
dist/local/server [options] <worldographer-file>
```

### Options

- `-host string`: Host to bind to (default: "localhost")
- `-port string`: Port to listen on (default: "8081") 
- `-timeout duration`: Automatically shutdown after this duration (default: 0, no timeout)
- `-h`: Show help

### Duration Format

The `-timeout` option accepts Go duration format:
- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `1h30m` - 1 hour 30 minutes
- `2h45m30s` - 2 hours 45 minutes 30 seconds

## Examples

### Basic Usage

```bash
# Start server with defaults (localhost:8081)
dist/local/server testdata/input/blank-2017-1.73-1.0.wxx

# Custom port
dist/local/server -port 9000 testdata/input/blank-2017-1.73-1.0.wxx

# Custom host and port
dist/local/server -host 127.0.0.1 -port 8888 testdata/input/blank-2017-1.73-1.0.wxx

# Bind to all interfaces
dist/local/server -host 0.0.0.0 -port 8080 testdata/input/blank-2017-1.73-1.0.wxx
```

### Auto-Shutdown Examples

```bash
# Auto-shutdown after 30 seconds
dist/local/server -timeout 30s testdata/input/blank-2017-1.73-1.0.wxx

# Auto-shutdown after 5 minutes with custom port
dist/local/server -port 9000 -timeout 5m testdata/input/blank-2017-1.73-1.0.wxx

# Testing configuration: shutdown after 1 minute
dist/local/server -host localhost -port 8888 -timeout 1m testdata/input/blank-2017-1.73-1.0.wxx
```

## Available Routes

### `GET /`
Main page displaying map information and hex grid preview.

**Response**: HTML page with map metadata and embedded SVG preview

### `GET /hex-grid.svg`
SVG visualization of the first 5x5 hex grid.

**Response**: SVG image showing hexagons with coordinate labels
- Each hex shows row,column coordinates (zero-based)
- Each hex shows cube coordinates using the format from `tile.Coords.String()`
- Hex geometry respects the map's orientation (flat-top vs pointy-top)

### `GET /ping`
Health check endpoint.

**Response**: Plain text "pong" with HTTP 200 status

### `GET /shutdown`
Gracefully shutdown the server.

**Response**: Plain text confirmation message, then server exits

## Testing

### Manual Testing

```bash
# Start server
dist/local/server -timeout 1m -port 9000 testdata/input/blank-2017-1.73-1.0.wxx &

# Wait for startup
sleep 2

# Health check
curl http://localhost:9000/ping

# View main page
curl http://localhost:9000/

# Get hex grid SVG
curl http://localhost:9000/hex-grid.svg

# Manual shutdown (or wait for timeout)
curl http://localhost:9000/shutdown
```

### Automated Testing Script

```bash
#!/bin/bash
set -e

PORT=9001
FILE="testdata/input/blank-2017-1.73-1.0.wxx"

echo "Starting server..."
dist/local/server -port $PORT -timeout 30s $FILE > server.log 2>&1 &

echo "Waiting for server startup..."
sleep 2

echo "Testing health endpoint..."
if curl -s http://localhost:$PORT/ping > /dev/null; then
    echo "✓ Server is healthy"
else
    echo "✗ Server health check failed"
    exit 1
fi

echo "Testing main page..."
if curl -s http://localhost:$PORT/ | grep -q "Worldographer Map Information"; then
    echo "✓ Main page loads correctly"
else
    echo "✗ Main page test failed"
    exit 1
fi

echo "Testing SVG endpoint..."
if curl -s http://localhost:$PORT/hex-grid.svg | grep -q "<svg"; then
    echo "✓ SVG endpoint works"
else
    echo "✗ SVG endpoint test failed"
    exit 1
fi

echo "Shutting down server..."
curl -s http://localhost:$PORT/shutdown > /dev/null

echo "✓ All tests passed!"
```

### Health Check Examples

```bash
# Simple health check
curl -s http://localhost:8081/ping

# Health check with HTTP status code
curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/ping

# Health check in shell script
if curl -s http://localhost:8081/ping > /dev/null 2>&1; then
    echo "Server is running"
else
    echo "Server is not responding"
fi

# Wait for server to be ready
while ! curl -s http://localhost:8081/ping > /dev/null 2>&1; do
    echo "Waiting for server..."
    sleep 1
done
echo "Server is ready!"
```

## Features

### Hex Grid Visualization
- Displays up to 5x5 grid of hexes from the map
- Shows proper hex geometry based on map orientation
- Labels each hex with:
  - Row,Column coordinates (zero-based)  
  - Cube coordinates in `+q+r+s` format
- Uses the map's configured hex width and height for accurate sizing

### Flexible Configuration
- Configurable host and port for different environments
- Optional auto-shutdown for testing and demos
- Health check endpoint for monitoring
- Manual shutdown endpoint for clean termination

### Multiple File Format Support
- Works with different Worldographer file versions
- Handles both 2017 and 2025 format files
- Graceful error handling for unsupported files

## Troubleshooting

### Common Issues

**Port already in use**
```bash
# Try a different port
dist/local/server -port 8082 your-file.wxx
```

**Can't connect to server**
```bash
# Check if server is running
curl http://localhost:8081/ping

# Check server logs if running in background
tail server.log
```

**File not found**
```bash
# Make sure file path is correct
ls -la testdata/input/
dist/local/server testdata/input/blank-2017-1.73-1.0.wxx
```

### Debugging

```bash
# Run in foreground to see logs
dist/local/server your-file.wxx

# Run in background and capture logs
dist/local/server your-file.wxx > server.log 2>&1 &
tail -f server.log
```
