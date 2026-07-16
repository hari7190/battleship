package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// 1. Serve the static assets directory
	// http.Dir searches relative to the directory where you RUN the application.
	// If you run it from the root directory, "ui/static" targets battleship/ui/static.
	fileServer := http.FileServer(http.Dir("ui/static"))

	// Strip the "/static/" prefix from the URL so that a request to "/static/css/style.css"
	// maps correctly inside the "ui/static" folder to "css/style.css".
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// 2. Serve the index.html file on the root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only match exactly "/" so it doesn't catch every unmatched route
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Reads directly from the disk path relative to application execution
		http.ServeFile(w, r, "ui/static/templates/index.html")
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
