package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

// --- Configuration Constants ---
const (
	caffeinePerCupMg = 95.0    // Average caffeine in a cup of coffee (mg)
	halfLifeHours    = 5.0     // Average half-life of caffeine in hours
	serverPort       = ":8080" // Port for the HTTP server
)

// CoffeeIntakeEvent stores the time and amount of a single coffee intake.
type CoffeeIntakeEvent struct {
	Time   time.Time `json:"time"`
	Amount float64   `json:"amount"`
}

// Tracker holds the state of coffee intake events.
// It's made thread-safe with a mutex for potential concurrent access in a real server.
type Tracker struct {
	mu     sync.Mutex
	events []CoffeeIntakeEvent
}

// NewTracker creates and returns a new Tracker instance.
func NewTracker() *Tracker {
	return &Tracker{
		events: make([]CoffeeIntakeEvent, 0),
	}
}

// AddCoffee logs a new coffee intake event with the current time.
func (t *Tracker) AddCoffee() {
	t.mu.Lock() // Lock to ensure thread safety when modifying events
	defer t.mu.Unlock()

	event := CoffeeIntakeEvent{
		Time:   time.Now(),
		Amount: caffeinePerCupMg,
	}
	t.events = append(t.events, event)
	fmt.Printf("Logged coffee at %s. Current count: %d\n", event.Time.Format("15:04:05"), len(t.events))
}

// CalculateCurrentCaffeineLevel calculates the total estimated caffeine in the system.
func (t *Tracker) CalculateCurrentCaffeineLevel() float64 {
	t.mu.Lock() // Lock for thread-safe access to events
	defer t.mu.Unlock()

	now := time.Now()
	totalCaffeine := 0.0

	if len(t.events) == 0 {
		return 0.0
	}

	for _, event := range t.events {
		timeElapsed := now.Sub(event.Time)
		timeElapsedHours := timeElapsed.Hours()

		// Only consider events that have already happened
		if timeElapsedHours < 0 {
			continue
		}

		// Caffeine decay formula: C = C0 * (0.5)^(t / T_half)
		remainingCaffeine := event.Amount * math.Pow(0.5, timeElapsedHours/halfLifeHours)
		totalCaffeine += remainingCaffeine
	}

	return totalCaffeine
}

// GetEvents returns all coffee intake events
func (t *Tracker) GetEvents() []CoffeeIntakeEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}

func main() {
	fmt.Println("--- Go Caffeine Tracker Backend Logic ---")
	tracker := NewTracker()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/add-coffee", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tracker.AddCoffee()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	http.HandleFunc("/api/caffeine-level", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		level := tracker.CalculateCurrentCaffeineLevel()
		json.NewEncoder(w).Encode(map[string]float64{"level": level})
	})

	http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		events := tracker.GetEvents()
		json.NewEncoder(w).Encode(events)
	})

	fmt.Printf("Server starting on http://localhost%s\n", serverPort)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
