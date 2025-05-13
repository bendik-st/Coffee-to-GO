# Go Caffeine Tracker

A simple web-based caffeine tracker built with Go for the backend and HTML/JS for the frontend.

## Features
- Log coffee intake with a single click
- See your current estimated caffeine level (mg)
- View your coffee intake history
- Modern, responsive UI

## How to Run

1. **Install Go**
   - Download and install Go from: https://golang.org/dl/

2. **Clone or Download this Repository**

3. **Run the Go Server**
   ```sh
   go run caffeine_tracker.go
   ```
   The server will start on [http://localhost:8080](http://localhost:8080)

4. **Open the App in Your Browser**
   - Go to: [http://localhost:8080](http://localhost:8080)
   - Use the web interface to add coffee and view your stats!

## Project Structure

- `caffeine_tracker.go` — Go backend with HTTP API
- `static/index.html` — Frontend HTML/JS/CSS

## API Endpoints
- `POST /api/add-coffee` — Log a new coffee
- `GET /api/caffeine-level` — Get current caffeine level
- `GET /api/events` — Get coffee intake history

---

Enjoy tracking your caffeine! ☕ 