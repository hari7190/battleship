package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Coordinate []int

type Placement struct {
	Ship      string       `json:"ship"`
	Positions []Coordinate `json:"positions"`
}

type Player struct {
	PlayerId string    `json:"player_id"`
	Position Placement `json:"position"`
}

type Game struct {
	GameId  string                 `json:"game_id"`
	Status  int8                   `json:"status_code"`
	Players map[string][]Placement `json:"players"`
}

type GameMemberShip struct {
	GameId   string `json:"game_id"`
	Token    string `json:"token"`
	PlayerId string `json:"player_id"`
}

type GameStore struct {
	Games map[string]Game
}

/*
	0 - game is created - no players
	1 - players joining
	2 - placements complete & ready to start
	3 - in progress
	4 - finished

*/

func NewGameStore() *GameStore {
	return &GameStore{
		Games: make(map[string]Game),
	}
}

// handle joining the game logic
func Join(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// bodyBytes, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	http.Error(w, "cant read data", http.StatusBadRequest)
		// 	return
		// }
		// defer r.Body.Close()

		game := Game{}
		var player_id string
		var gameMemberShip GameMemberShip

		// if the token exists, get the ids, check if game exists, retrieve it.
		if r.Header.Get("token") != "null" {
			tokenParts := strings.Split(r.Header.Get("token"), ":")
			incomingGameId := tokenParts[0]
			incomingPlayerId := tokenParts[1]

			var exists bool
			game, exists = store.Games[incomingGameId]

			if exists {
				log.Default().Println("Game exists")
				player_id = incomingPlayerId
				gameMemberShip = GameMemberShip{
					GameId:   incomingGameId,
					PlayerId: incomingPlayerId,
					Token:    r.Header.Get("token"),
				}
			} else {
				gameMemberShip = store.createGame(game, player_id)
			}
		} else { // else create a new game
			gameMemberShip = store.createGame(game, player_id)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(gameMemberShip); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

func (store *GameStore) createGame(game Game, player_id string) GameMemberShip {
	game = Game{
		GameId:  uuid.NewString(),
		Players: make(map[string][]Placement),
	}
	player_id = uuid.NewString()
	game.Players[player_id] = []Placement{}
	token := game.GameId + ":" + uuid.NewString()
	store.addGame(game)

	gameMemberShip := GameMemberShip{
		GameId:   game.GameId,
		Token:    token,
		PlayerId: player_id,
	}

	log.Default().Println("Game created: " + game.GameId)
	log.Default().Println("Player joined:" + player_id)

	return gameMemberShip
}

// handle ship placement events
func Place(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		if token == "" {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(token, ":")

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var placement Placement
		if err := json.Unmarshal(bodyBytes, &placement); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// gameID, err := IsJoinableGameAvailable()
		if err != nil {
			log.Printf("failed to read game id: %v", err)
			http.Error(w, "failed to save placement", http.StatusInternalServerError)
			return
		}

		store.updatePlacement(placement, tokenParts[1], tokenParts[0])
		w.WriteHeader(http.StatusCreated)
	}
}

func IsJoinableGameAvailable() (Game, error) {
	// iterate
}

func (gs *GameStore) addGame(game Game) {
	gs.Games[game.GameId] = game
}

func (gs *GameStore) updatePlacement(placement Placement, player_id string, gameId string) {
	game, _ := gs.Games[gameId]

	game.Players[player_id] = append(game.Players[player_id], placement)
}
