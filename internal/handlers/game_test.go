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
