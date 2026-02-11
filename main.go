package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// Tweak as needed
type NowPlayingData struct {
	SongName   string `json:"songName"`
	ArtistName string `json:"artistName"`
	AlbumName  string `json:"albumName"`
}

var (
	// Store the global now playing state
	currentTrack NowPlayingData

	// And a mutex to access it
	mu sync.RWMutex

	// Store each connection
	conns map[chan string]struct{}

	// And a mutex to add/remove them
	connsMu sync.RWMutex
)

func updateHandler(w http.ResponseWriter, r *http.Request) {
	// Check system PSK
	psk := os.Getenv("SYSTEM_PSK")
	if psk == "" {
		log.Fatal("SYSTEM_PSK must be set. Currently empty.")
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey != psk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var nextTrack NowPlayingData
	if err := json.Unmarshal(body, &nextTrack); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	currentTrack = nextTrack
	mu.Unlock()

	log.Println("Received new track!")
	log.Printf(
		"%s by %s (%s)\n",
		currentTrack.SongName,
		currentTrack.ArtistName,
		currentTrack.AlbumName,
	)

	// send
	data, _ := json.Marshal(nextTrack)
	msg := fmt.Sprintf("data: %s\n\n", data)
	connsMu.RLock()
	for conn, _ := range conns {
		conn <- msg
	}
	connsMu.RUnlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated"))
}

func nowPlayingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	conn := make(chan string, 8)

	// This is evil Go for
	// set.add(...) + mutex
	connsMu.Lock()
	conns[conn] = struct{}{}
	connsMu.Unlock()

	mu.RLock()
	track := currentTrack
	mu.RUnlock()

	data, _ := json.Marshal(track)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()

	for {
		select {
		case msg := <-conn:
			fmt.Fprintf(w, msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			connsMu.Lock()
			delete(conns, conn)
			connsMu.Unlock()
			return
		}
	}
}

func main() {
	godotenv.Load()

	currentTrack = NowPlayingData{
		SongName:   "Now Playing",
		ArtistName: "",
		AlbumName:  "",
	}

	conns = make(map[chan string]struct{})

	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/now-playing", nowPlayingHandler)

	log.Println("Server starting on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalln(err)
	}
}
