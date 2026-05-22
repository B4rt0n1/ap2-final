package redis

import (
	"context"
	"encoding/json"
	"time"

	"inventory-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type redisCarCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCarCache(client *redis.Client, ttl time.Duration) domain.CarCache {
	return &redisCarCache{
		client: client,
		ttl:    ttl,
	}
}

func (r *redisCarCache) Get(ctx context.Context, id string) (*domain.Car, error) {
	val, err := r.client.Get(ctx, "car:"+id).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss safely returned
	} else if err != nil {
		return nil, err
	}

	var car domain.Car
	if err := json.Unmarshal([]byte(val), &car); err != nil {
		return nil, err
	}
	return &car, nil
}

func (r *redisCarCache) Set(ctx context.Context, car *domain.Car) error {
	data, err := json.Marshal(car)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "car:"+car.ID, data, r.ttl).Err()
}

func (r *redisCarCache) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, "car:"+id).Err()
}
