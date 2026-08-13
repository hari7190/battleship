package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// weird type to accomodate for future cell information
type Coordinate struct {
	X   int  `json:"x"`
	Y   int  `json:"y"`
	Hit bool `json:"hit"`
}

type Placement struct {
	Ship      string       `json:"ship"`
	Positions []Coordinate `json:"positions"`
}

// type Player struct {
// 	PlayerId     string         `json:"player_id"`
// 	Position     Placement      `json:"position"`
// 	FiringRecord []FiringRecord `json:"firingrecord"`
// }

type Game struct {
	GameId     string                  `json:"game_id"`
	Status     int8                    `json:"status_code"`
	Players    map[string][]Placement  `json:"players"`
	Misses     map[string][]Coordinate `json:"misses"`
	NextPlayer string                  `json:"next_player"`
	Winner     string                  `json:"winner"`
}

type PlayerBoardState struct {
	Ships  []Placement  `json:"ships"`
	Misses []Coordinate `json:"misses"`
}

type GameMemberShip struct {
	GameId             string `json:"game_id"`
	Token              string `json:"token"`
	PlayerId           string `json:"player_id"`
	WaitingForOpponent bool   `json:"waiting_for_opponent"`
}

type GameStore struct {
	Games       map[string]Game
	subscribers map[string]map[string][]chan string
}

type FiringRecord struct {
	X       int  `json:"x"`
	Y       int  `json:"y"`
	SUCCESS bool `json:"success"`
}

func NewGameStore() *GameStore {
	return &GameStore{
		Games:       make(map[string]Game),
		subscribers: make(map[string]map[string][]chan string),
	}
}

func (gs *GameStore) AddSubscriber(gameID string, playerID string, ch chan string) {
	if gs.subscribers[gameID] == nil {
		gs.subscribers[gameID] = make(map[string][]chan string)
	}
	gs.subscribers[gameID][playerID] = append(gs.subscribers[gameID][playerID], ch)

}

func (gs *GameStore) RemoveSubscriber(gameID string, playerID string, ch chan string) {
	subs := gs.subscribers[gameID][playerID]
	for i, c := range subs {
		if c == ch {
			gs.subscribers[gameID][playerID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

func (gs *GameStore) Broadcast(gameID string, playerID string, payload string) {
	if gs.subscribers[gameID] == nil {
		return
	}
	for _, ch := range gs.subscribers[gameID][playerID] {
		select {
		case ch <- payload:
		default:
		}
	}
}

func (gs *GameStore) BroadcastAll(gameID string, payload string) {
	if gs.subscribers[gameID] == nil {
		return
	}
	for playerID := range gs.subscribers[gameID] {
		gs.Broadcast(gameID, playerID, payload)
	}
}

func parseGameToken(token string) (gameID string, playerID string, ok bool) {
	if token == "" || token == "null" {
		return "", "", false
	}
	gameID, playerID, found := strings.Cut(token, ":")
	if !found || gameID == "" || playerID == "" {
		return "", "", false
	}
	return gameID, playerID, true
}

func allShipsSunk(placements []Placement) bool {
	if len(placements) == 0 {
		return false
	}
	for _, placement := range placements {
		if len(placement.Positions) == 0 {
			return false
		}
		for _, pos := range placement.Positions {
			if !pos.Hit {
				return false
			}
		}
	}
	return true
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
		if incomingGameId, incomingPlayerId, ok := parseGameToken(r.Header.Get("token")); ok {
			var exists bool
			game, exists = store.Games[incomingGameId]

			if exists {
				log.Default().Println("Game exists")
				player_id = incomingPlayerId
				gameMemberShip = GameMemberShip{
					GameId:             incomingGameId,
					PlayerId:           incomingPlayerId,
					Token:              r.Header.Get("token"),
					WaitingForOpponent: len(game.Players) < 2,
				}
			} else {
				gameMemberShip = store.createGame(game, player_id)
			}
		} else { // else create a new game
			game, err := store.IsJoinableGameAvailable()
			if err != nil {
				log.Default().Println("Error - finding a new game")
			}

			if game.GameId != "" {
				player_id := uuid.NewString()
				game, err = store.addPlayerToGame(game.GameId, player_id)

				gameMemberShip = GameMemberShip{
					GameId:             game.GameId,
					Token:              game.GameId + ":" + player_id,
					PlayerId:           player_id,
					WaitingForOpponent: false,
				}
			} else {
				gameMemberShip = store.createGame(game, player_id)
			}

		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(gameMemberShip); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}
}

// handle ship placement events
func Place(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId, playerId, ok := parseGameToken(r.Header.Get("token"))
		if !ok {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}

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

		store.updatePlacement(placement, playerId, gameId)
		w.WriteHeader(http.StatusCreated)
	}
}

// handle retrieving Game Data
func GetPlayerData(store *GameStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId, playerId, ok := parseGameToken(r.Header.Get("token"))
		if !ok {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}
		err := json.NewEncoder(w).Encode(store.playerBoardState(gameId, playerId))
		if err != nil {
			log.Printf("Failed to jsonize: %v", err)
		}
	}
}

func (store *GameStore) playerBoardState(game_id string, player_id string) PlayerBoardState {
	game := store.Games[game_id]
	misses := game.Misses[player_id]
	if misses == nil {
		misses = []Coordinate{}
	}
	ships := game.Players[player_id]
	if ships == nil {
		ships = []Placement{}
	}
	return PlayerBoardState{
		Ships:  ships,
		Misses: misses,
	}
}

func GetPlayerDataFromStore(store *GameStore, game_id string, player_id string) string {
	jsonBytes, err := json.Marshal(store.playerBoardState(game_id, player_id))
	if err != nil {
		log.Printf("Failed to jsonize: %v", err)
	}
	return string(jsonBytes)
}

func (store *GameStore) IsJoinableGameAvailable() (Game, error) {
	for game := range store.Games {
		if len(store.Games[game].Players) < 2 {
			return store.Games[game], nil
		}
	}
	return Game{}, nil
}

func (store *GameStore) createGame(game Game, player_id string) GameMemberShip {
	game = Game{
		GameId:  uuid.NewString(),
		Players: make(map[string][]Placement),
		Misses:  make(map[string][]Coordinate),
	}
	player_id = uuid.NewString()
	game.Players[player_id] = []Placement{}
	game.Misses[player_id] = []Coordinate{}
	token := game.GameId + ":" + player_id
	store.addGameToStore(game)

	gameMemberShip := GameMemberShip{
		GameId:             game.GameId,
		Token:              token,
		PlayerId:           player_id,
		WaitingForOpponent: true,
	}

	log.Default().Println("Game created: " + game.GameId)
	log.Default().Println("Player joined:" + player_id)

	return gameMemberShip
}

func (gs *GameStore) addGameToStore(game Game) {
	gs.Games[game.GameId] = game
}

func (gs *GameStore) addPlayerToGame(gameId string, player_id string) (Game, error) {
	game, exists := gs.Games[gameId]

	if exists {
		if len(game.Players) < 2 {
			game.Players[player_id] = []Placement{}
			if game.Misses == nil {
				game.Misses = make(map[string][]Coordinate)
			}
			game.Misses[player_id] = []Coordinate{}
			gs.Games[gameId] = game
			log.Default().Println("Player joined:" + player_id)

			// Notify the waiting player that an opponent joined
			for existingPlayerId := range game.Players {
				if existingPlayerId != player_id {
					gs.Broadcast(gameId, existingPlayerId, "player-joined>:true")
				}
			}

			return game, nil
		}
	}
	return Game{}, errors.New("Failed to add to game")
}

func (gs *GameStore) updatePlacement(placement Placement, player_id string, gameId string) {
	game, _ := gs.Games[gameId]

	game.Players[player_id] = append(game.Players[player_id], placement)

	//if both players have 3 placements, set fire-control to false
	//iterate over the players and check if they have 3 placements each.
	for _, player := range game.Players {
		if len(player) != 3 {
			return
		}
	}
	if len(game.Players) == 2 {
		for playerId := range game.Players {
			gs.Broadcast(gameId, playerId, "game-ready>:true")
			gs.Broadcast(gameId, playerId, "fire-control>:"+"false")
		}
	}
}
