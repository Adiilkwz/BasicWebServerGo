# Assignment 2: Concurrent Web Server in Go

This project is a thread-safe, concurrent web server built using Go's standard `net/http` package. It implements a basic Key-Value store with RESTful endpoints, background monitoring, and graceful shutdown capabilities.

## Features

* **REST API:** Endpoints to Create, Read, and Delete data.
* **Thread Safety:** Uses `sync.Mutex` to safely manage shared state (map and counters) across concurrent HTTP requests.
* **Background Worker:** A concurrent Goroutine that logs server statistics (Request count & Database size) every 5 seconds.
* **Graceful Shutdown:** Captures OS signals (`Ctrl+C`) to shut down background workers and exit cleanly.
* **Routing:** Utilizes Go 1.22+ `http.NewServeMux` for pattern matching (e.g., `DELETE /data/{key}`).

## specific Requirements Implemented

1.  **Concurrency:** `net/http` handles requests concurrently; Mutexes protect the critical sections.
2.  **Coordination:** Channels are used to signal the background worker to stop.
3.  **Multiplexing:** `select-case` handles the periodic ticker and shutdown signals.

---

## How to Run

1.  **Prerequisites:** Ensure you have Go installed (Version 1.22 or higher is required for the URL path matching).
2.  **Run the Server:**
    ```bash
    go run main.go
    ```
3.  **Output:** You will see:
    ```text
    Server starting on :8000
    ```

---

## API Endpoints

The server runs on **port 8000**.

### 1. Store Data (POST)
Adds a key-value pair to the server.

* **URL:** `/data`
* **Method:** `POST`
* **Body:** JSON
    ```json
    {
      "key": "course",
      "value": "Advanced Programming"
    }
    ```
* **Command:**
    ```bash
    curl -X POST http://localhost:8000/data \
         -H "Content-Type: application/json" \
         -d '{"key": "user1", "value": "Alice"}'
    ```

### 2. Retrieve All Data (GET)
Returns the entire content of the in-memory map.

* **URL:** `/data`
* **Method:** `GET`
* **Command:**
    ```bash
    curl http://localhost:8000/data
    ```

### 3. Delete Data (DELETE)
Removes a specific item by its key.

* **URL:** `/data/{key}`
* **Method:** `DELETE`
* **Command:**
    ```bash
    curl -X DELETE http://localhost:8000/data/user1
    ```

### 4. Server Statistics (GET)
Returns the total number of requests processed.

* **URL:** `/stats`
* **Method:** `GET`
* **Command:**
    ```bash
    curl http://localhost:8000/stats
    ```

---

## Internal Architecture

### The `Server` Struct
The core state is managed within the `Server` struct:
```go
type Server struct {
    mu         sync.Mutex        // Protects data and requests
    data       map[string]string // In-memory database
    requests   int               // Total request counter
    shutDownCh chan struct{}     // Signal channel for shutdown
}