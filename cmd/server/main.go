package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type TelemetryEvent struct {
	DeviceID  string  `json:"device_id"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

type RecentEvents struct {
	mu     sync.Mutex
	limit  int
	events []TelemetryEvent
}

func NewRecentEvents(limit int) *RecentEvents {
	return &RecentEvents{limit: limit}
}

func (r *RecentEvents) Add(event TelemetryEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
	if len(r.events) > r.limit {
		r.events = r.events[len(r.events)-r.limit:]
	}
}

func (r *RecentEvents) List() []TelemetryEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	events := make([]TelemetryEvent, len(r.events))
	copy(events, r.events)
	return events
}

func main() {
	eventQueue := make(chan TelemetryEvent, 1000)
	recentEvents := NewRecentEvents(20)

	startWorkers(eventQueue, 4)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/events", eventsHandler(eventQueue, recentEvents))
	http.HandleFunc("/events/recent", recentEventsHandler(recentEvents))

	fmt.Println("Downlink server listening on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "ERROR: Must use GET method.", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func eventsHandler(eventQueue chan<- TelemetryEvent, recentEvents *RecentEvents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "ERROR: Must use POST method.", http.StatusMethodNotAllowed)
			return
		}

		var event TelemetryEvent

		err := json.NewDecoder(r.Body).Decode(&event)

		if err != nil {
			http.Error(w, "Invalid JSON.", http.StatusBadRequest)
			return
		}

		if event.DeviceID == "" || event.Metric == "" || event.Unit == "" {
			http.Error(w, "JSON is missing required fields.", http.StatusBadRequest)
			return
		}

		if event.Timestamp == "" {
			event.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		select {
		case eventQueue <- event:
			recentEvents.Add(event)
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Event accepted."))
		default:
			http.Error(w, "Event queue full.", http.StatusServiceUnavailable)
		}
	}
}

func recentEventsHandler(recentEvents *RecentEvents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "ERROR: Must use GET method.", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recentEvents.List())
	}
}

func startWorkers(eventQueue <-chan TelemetryEvent, workerCount int) {
	for i := 1; i <= workerCount; i++ {
		go func(workerID int) {
			for event := range eventQueue {
				fmt.Printf(
					"worker=%d processed device=%s metric=%s value=%.2f unit=%s timestamp=%s\n",
					workerID,
					event.DeviceID,
					event.Metric,
					event.Value,
					event.Unit,
					event.Timestamp,
				)
			}
		}(i)
	}
}
