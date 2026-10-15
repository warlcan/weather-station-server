package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *RedisRepository

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func InitRedis(addr string, password string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	Redis = &RedisRepository{
		client: rdb,
		ctx:    ctx,
	}

	return nil
}

func (r *RedisRepository) GetString(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisRepository) ExistsString(key string) (int64, error) {
	val, err := r.client.Exists(r.ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *RedisRepository) WriteStringTTL(key string, value string, ttl int) error {
	err := r.client.Set(r.ctx, key, value, time.Duration(ttl)*time.Minute).Err()
	return err
}
