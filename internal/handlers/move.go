package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Placement struct {
	Ship      string       `json:"ship"`
	Positions []Coordinate `json:"positions"`
}

type Player struct {
	PlayerId string    `json:"player_id"`
	Position Placement `json:"position"`
}

type Coordinate []int

type Game struct {
	GameId  string `json:"game_id"`
	Players []Player
}

// type: "blue", positions: [[2, 4], [2, 5], [2, 6], [2, 7]]
func Place() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data := Game{
			GameId: "game_1223",
			Players: []Player{
				{
					PlayerId: "1232",
					Position: Placement{
						Ship: "blue",
						Positions: []Coordinate{
							{2, 4},
							{2, 5},
							{2, 6},
							{2, 7},
						},
					},
				}, {
					PlayerId: "12342",
					Position: Placement{
						Ship: "orange",
						Positions: []Coordinate{
							{2, 4},
							{2, 5},
							{2, 6},
							{2, 7},
						},
					},
				},
			},
		}
		bytes, _ := json.MarshalIndent(data, "", "  ")
		// tm := time.Now().Format(time.RFC1123)
		// w.Write(bytes)
		err := os.WriteFile(data.GameId+".json", bytes, 0644)
		if err != nil {
			log.Fatalf("Error writing file: %v", err)
		}
		log.Println("JSON file written successfully!")
	}
}
