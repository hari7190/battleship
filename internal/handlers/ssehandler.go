package handlers

import (
	"log"
	"net/http"
	"strings"
)

func SSEHandler(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId, playerId, ok := parseGameToken(r.PathValue("token"))
		if !ok {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}

		ch := make(chan string, 8)
		store.AddSubscriber(gameId, playerId, ch)
		defer store.RemoveSubscriber(gameId, playerId, ch)

		// 1. Set SSE Headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 2. Ensure ResponseWriter supports flushing
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// 3. Monitor client disconnection using Request Context
		ctx := r.Context()

		log.Println("Client connected to SSE stream")

		for {
			select {
			case msg := <-ch:
				eventType, data, found := strings.Cut(msg, ">:")
				if !found {
					log.Printf("ignoring malformed SSE payload: %q", msg)
					continue
				}
				_, _ = w.Write([]byte("event: " + eventType + "\ndata: " + data + "\n\n"))
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}
	}
}
