package main

import (
	"context"
	"log"
	"net/http"

	"hari.foo/battleship/internal/handlers"
)

func main() {
	store := handlers.NewGameStore()
	context := context.Background()
	redisStore := handlers.NewKVStore("localhost:6379")

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("ui/static"))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /api/playerInfo", handlers.GetPlayerData(store))
	mux.Handle("GET /api/events/{token}", handlers.SSEHandler(store))

	mux.Handle("POST /api/join", handlers.Join(store, context, redisStore))
	mux.Handle("POST /api/placement", handlers.Place(store))
	mux.Handle("POST /api/fire", handlers.Fire(store))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "ui/static/templates/index.html")
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
