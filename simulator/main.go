package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const eventsURL = "http://localhost:8080/events"

type TelemetryEvent struct {
	DeviceID  string  `json:"device_id"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

type SensorMetric struct {
	Metric string
	Unit   string
	Min    float64
	Max    float64
}

var metrics = []SensorMetric{
	{Metric: "altitude", Unit: "ft", Min: 10000, Max: 42000},
	{Metric: "airspeed", Unit: "knots", Min: 120, Max: 560},
	{Metric: "engine_temp", Unit: "c", Min: 650, Max: 950},
	{Metric: "fuel_level", Unit: "percent", Min: 0, Max: 100},
	{Metric: "vertical_speed", Unit: "ft/min", Min: -3000, Max: 3000},
}

func main() {
	rand.Seed(time.Now().UnixNano())

	for {
		event := createRandomEvent()
		if err := sendEvent(event); err != nil {
			fmt.Println("Event failed:", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func createRandomEvent() TelemetryEvent {
	aircraftID := rand.Intn(10) + 1
	selectedMetric := metrics[rand.Intn(len(metrics))]

	return TelemetryEvent{
		DeviceID:  fmt.Sprintf("aircraft-%03d", aircraftID),
		Metric:    selectedMetric.Metric,
		Value:     selectedMetric.Min + rand.Float64()*(selectedMetric.Max-selectedMetric.Min),
		Unit:      selectedMetric.Unit,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func sendEvent(event TelemetryEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(
		eventsURL,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Event sent, server returned:", resp.Status)
	return nil
}
