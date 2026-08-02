package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

		for playerId, placements := range game.Players {
			if currentPlayerId == playerId {
				continue
			}

			for i, placement := range placements {
				for j, pos := range placement.Positions {
					if pos.X == cell.X && pos.Y == cell.Y {
						log.Default().Printf("%s's Ship %s HIT \n", playerId, placement.Ship)
						game.Players[playerId][i].Positions[j].Hit = true
						hit = true
						break
					}
				}
				if hit {
					break
				}
			}

			if !hit {
				if game.Misses == nil {
					game.Misses = make(map[string][]Coordinate)
				}
				game.Misses[playerId] = append(game.Misses[playerId], Coordinate{X: cell.X, Y: cell.Y})
				log.Default().Printf("Miss on %s at (%d,%d)\n", playerId, cell.X, cell.Y)
			}

			gs.Games[gameId] = game
			gs.Broadcast(gameId, playerId, "update>:"+GetPlayerDataFromStore(gs, gameId, playerId))
			gs.Broadcast(gameId, playerId, "fire-control>:"+"false")
		}

		if hit {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}
}
