package handlers

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type KVStore struct {
	rdb *redis.Client
}

func NewKVStore(addr string) *KVStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &KVStore{rdb: rdb}
}

func (store *KVStore) Set(ctx context.Context, key, value string) error {
	err := store.rdb.Set(ctx, key, value, 0).Err()

	if err != nil {
		return fmt.Errorf("Error trying to set key %w", err)
	}

	return nil
}

func (store *KVStore) Get(ctx context.Context, key string) (string, error) {
	value, err := store.rdb.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", fmt.Errorf("Key not found : %w", err)
	} else if err != nil {
		return "", fmt.Errorf("Unable to retrieve: %w", err)
	}

	return value, nil
}
