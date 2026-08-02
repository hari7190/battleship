package handlers

import "testing"

func TestAddSubscriberAndBroadcast(t *testing.T) {
	store := NewGameStore()
	ch := make(chan string, 1)

	store.AddSubscriber("game-1", "player-1", ch)

	if got := len(store.subscribers["game-1"]["player-1"]); got != 1 {
		t.Fatalf("expected 1 subscriber, got %d", got)
	}

	store.Broadcast("game-1", "player-1", "payload")

	select {
	case msg := <-ch:
		if msg != "payload" {
			t.Fatalf("expected payload, got %q", msg)
		}
	default:
		t.Fatal("expected broadcast payload")
	}
}

func TestAddPlayerToGameBroadcastsPlayerJoined(t *testing.T) {
	store := NewGameStore()
	game := Game{
		GameId:  "game-1",
		Players: map[string][]Placement{"player-1": {}},
	}
	store.addGameToStore(game)

	ch := make(chan string, 1)
	store.AddSubscriber("game-1", "player-1", ch)

	_, err := store.addPlayerToGame("game-1", "player-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case msg := <-ch:
		if msg != "player-joined>:true" {
			t.Fatalf("expected player-joined event, got %q", msg)
		}
	default:
		t.Fatal("expected player-joined broadcast to waiting player")
	}
}

func TestAllShipsSunk(t *testing.T) {
	if allShipsSunk(nil) {
		t.Fatal("empty placements should not be sunk")
	}

	partial := []Placement{{
		Ship: "blue",
		Positions: []Coordinate{
			{X: 1, Y: 1, Hit: true},
			{X: 2, Y: 1, Hit: false},
		},
	}}
	if allShipsSunk(partial) {
		t.Fatal("partial hits should not be sunk")
	}

	sunk := []Placement{
		{Ship: "blue", Positions: []Coordinate{{X: 1, Y: 1, Hit: true}, {X: 2, Y: 1, Hit: true}}},
		{Ship: "green", Positions: []Coordinate{{X: 3, Y: 3, Hit: true}}},
	}
	if !allShipsSunk(sunk) {
		t.Fatal("expected all ships sunk")
	}
}
