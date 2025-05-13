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
	serverPort = ":8080" // Port for the HTTP server
)

// CoffeeIntakeEvent stores the time and amount of a single coffee intake.
type CoffeeIntakeEvent struct {
	Time   time.Time `json:"time"`
	Amount float64   `json:"amount"`
}

// DrinkRequest represents the incoming request to add a drink
type DrinkRequest struct {
	Amount float64 `json:"amount"`
}

// ForecastPoint represents a point in time with predicted caffeine level
type ForecastPoint struct {
	Time        time.Time `json:"time"`
	Caffeine    float64   `json:"caffeine"`
	HasDrink    bool      `json:"hasDrink"`
	DrinkAmount float64   `json:"drinkAmount,omitempty"`
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

// AddDrink logs a new drink intake event with the current time and specified amount.
func (t *Tracker) AddDrink(amount float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := CoffeeIntakeEvent{
		Time:   time.Now(),
		Amount: amount,
	}
	t.events = append(t.events, event)
	fmt.Printf("Logged drink at %s. Current count: %d\n", event.Time.Format("15:04:05"), len(t.events))
}

// CalculateCaffeineLevelAt calculates the caffeine level at a specific time
func (t *Tracker) CalculateCaffeineLevelAt(targetTime time.Time) float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	totalCaffeine := 0.0

	if len(t.events) == 0 {
		return 0.0
	}

	for _, event := range t.events {
		timeElapsed := targetTime.Sub(event.Time)
		timeElapsedHours := timeElapsed.Hours()

		if timeElapsedHours < 0 {
			continue
		}

		// Caffeine decay formula: C = C0 * (0.5)^(t / T_half)
		remainingCaffeine := event.Amount * math.Pow(0.5, timeElapsedHours/5.0) // 5 hours half-life
		totalCaffeine += remainingCaffeine
	}

	return totalCaffeine
}

// GenerateForecast generates a forecast of caffeine levels for the next 24 hours
func (t *Tracker) GenerateForecast() []ForecastPoint {
	now := time.Now()
	forecast := make([]ForecastPoint, 0)

	// Generate points for every 30 minutes for the next 24 hours
	for i := 0; i < 48; i++ {
		targetTime := now.Add(time.Duration(i*30) * time.Minute)
		caffeine := t.CalculateCaffeineLevelAt(targetTime)

		// Check if there's a drink at this time
		var hasDrink bool
		var drinkAmount float64
		for _, event := range t.events {
			if event.Time.Format("15:04") == targetTime.Format("15:04") {
				hasDrink = true
				drinkAmount = event.Amount
				break
			}
		}

		forecast = append(forecast, ForecastPoint{
			Time:        targetTime,
			Caffeine:    caffeine,
			HasDrink:    hasDrink,
			DrinkAmount: drinkAmount,
		})
	}

	return forecast
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

		var req DrinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		tracker.AddDrink(req.Amount)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	http.HandleFunc("/api/caffeine-level", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		level := tracker.CalculateCaffeineLevelAt(time.Now())
		json.NewEncoder(w).Encode(map[string]float64{"level": level})
	})

	http.HandleFunc("/api/forecast", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		forecast := tracker.GenerateForecast()
		json.NewEncoder(w).Encode(forecast)
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
