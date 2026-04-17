package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(url string) *Redis {
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("Invalid REDIS_URL, caching disabled: %v", err)
		return nil
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed, caching disabled: %v", err)
		return nil
	}

	log.Println("Connected to Redis")
	return &Redis{client: client}
}

func (r *Redis) Get(ctx context.Context, key string, dest interface{}) bool {
	if r == nil {
		return false
	}
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(val), dest) == nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if r == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	r.client.Set(ctx, key, data, ttl)
}

func (r *Redis) Delete(ctx context.Context, keys ...string) {
	if r == nil {
		return
	}
	r.client.Del(ctx, keys...)
}

func (r *Redis) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if r == nil {
		return 0, nil
	}
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (r *Redis) Close() {
	if r == nil {
		return
	}
	r.client.Close()
}
