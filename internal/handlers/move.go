package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
	GameId  string `json:"game_id"`
	Players []Player
}

func Place() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatalf("Cannot read body", err)
		}

		defer r.Body.Close()

		bodyString := string(bodyBytes)

		log.Println(bodyString)
		var placement Placement
		err = json.Unmarshal(bodyBytes, &placement)
		

		log.Println(placement)

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
		data.Players[0].Position = placement


		bytes, _ := json.MarshalIndent(data, "", "  ")
		// tm := time.Now().Format(time.RFC1123)
		// w.Write(bytes)
		err = os.WriteFile(data.GameId+".json", bytes, 0644)
		if err != nil {
			log.Fatalf("Error writing file: %v", err)
		}

		WriteToCache(data)

		log.Println("JSON file written successfully!")
	}
}

func WriteToCache(game Game) error {
	jsonData, err := json.Marshal(game)
	if err != nil {
		return err
	}
	jsonString := string(jsonData)

	conn, err := net.Dial("tcp", "127.0.0.1:6379")

	if err != nil {
		return err
	}

	defer conn.Close()

	// Format: *3\r\n$3\r\nSET\r\n$<key_len>\r\n<key>\r\n$<val_len>\r\n<val>\r\n
	respCommand := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(game.GameId), game.GameId, len(jsonString), jsonString)

	_, err = conn.Write([]byte (respCommand))
	
	if err != nil{
		return err
	}
	return nil
}	