# Downlink

Downlink is a small Go backend for ingesting simulated telemetry events over HTTP.

It includes a server that receives JSON telemetry data and a simulator that generates random events and POSTs them to the server. The current MVP stores the latest 20 events in memory.

## Features

* Go HTTP server
* Telemetry simulator
* `POST /events` ingestion endpoint
* Endpoint to retrieve latest events
* In-memory recent event storage
* Mutex-protected shared state

## Project Structure

```text
downlink/
  cmd/
    server/       # HTTP server
    simulator/    # fake telemetry generator
  internal/       # shared app code
  go.mod
```

## Run

Start the server:

```bash
go run ./cmd/server
```

In another terminal, start the simulator:

```bash
go run ./cmd/simulator
```

The server runs at:

```text
http://localhost:8080
```

## API

Health check:

```bash
curl http://localhost:8080/health
```

Send an event:

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "sensor-001",
    "metric": "temperature_celsius",
    "value": 87.2,
    "unit": "celsius"
  }'
```

Retrieve latest events:

```bash
curl http://localhost:8080/events
```

## Roadmap

* Add async worker pool
* Add channel-based event queue
* Persist events with PostgreSQL
* Add alert rules
* Add metrics for throughput and queue depth
* Add Docker Compose setup

## Status

Early MVP. The current focus is understanding the basic event ingestion flow before adding persistence, queueing, and observability.
