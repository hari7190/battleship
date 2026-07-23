package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)

func Fire(gs *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//get token
		token := r.Header.Get("token")
		tokenParts := strings.Split(token, ":")
		bodyBytes, err := io.ReadAll(r.Body)
		var hit bool
		if err != nil {
			http.Error(w, "cant read data", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var cell Coordinate
		if err := json.Unmarshal(bodyBytes, &cell); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		gameId := tokenParts[0]
		currentPlayerId := tokenParts[1]

		game := gs.Games[gameId]

		for playerId, postions := range game.Players {
			if currentPlayerId != playerId {
				for _, placement := range postions {
					ship := placement.Ship
					positions := placement.Positions
					if slices.Contains(positions, cell) {
						log.Default().Printf("Ship %s HIT \n", ship)
						index := slices.Index(positions, cell)
						positions[index].Hit = true
						hit = true
						break
					}
				}
			}
		}
		if hit {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}
}
