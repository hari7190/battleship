package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func SSEHandler(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenParts := strings.Split(r.PathValue("token"), ":")
		playerId := tokenParts[1]
		gameId := tokenParts[0]

		ch := make(chan string, 8)
		store.AddSubscriber(gameId, playerId, ch)
		defer store.RemoveSubscriber(gameId, playerId, ch)

		// 1. Set SSE Headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// Header for CORS if your frontend is on a different origin/port
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

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// eventID := 0
		for {
			select {
			case msg := <-ch:
				msgParts := strings.Split(msg, ">:")
				eventType := msgParts[0]
				data := msgParts[1]
				_, _ = w.Write([]byte("event: " + eventType + "\ndata: " + data + "\n\n"))
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}
	}
}
