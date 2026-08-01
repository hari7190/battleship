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

		ch := make(chan string, 1)
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
				_, _ = w.Write([]byte("event: update\ndata: " + msg + "\n\n"))
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}

		// for {
		// 	select {
		// 	case <-ctx.Done():
		// 		// Client closed the connection or timed out
		// 		log.Println("Client disconnected from SSE stream")
		// 		return

		// 	case t := <-ticker.C:
		// 		eventID++

		// 		// 4. Format the SSE message standard format:
		// 		// "id: <id>\nevent: <name>\ndata: <content>\n\n"
		// 		message2 := GetPlayerDataFromStore(store, gameId, playerId)
		// 		message := fmt.Sprintf("id: %d\nevent: ping\ndata: %s\n\n", eventID, message2)
		// 		fmt.Print(t.Format(time.RFC3339))
		// 		fmt.Print(message2)
		// 		// Write message to buffer
		// 		_, err := w.Write([]byte(message))
		// 		if err != nil {
		// 			log.Printf("Error writing to stream: %v\n", err)
		// 			return
		// 		}

		// 		// Flush buffer directly to the wire
		// 		flusher.Flush()
		// 	}
		// }
	}
}
