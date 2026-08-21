package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(url string) *Redis {
	opts, err := redis.ParseURL(url)
	if err != nil {
		slog.Warn("invalid REDIS_URL, caching disabled", "error", err)
		return nil
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Warn("redis connection failed, caching disabled", "error", err)
		return nil
	}

	slog.Info("connected to redis")
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

var incrementWithTTLScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func (r *Redis) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if r == nil {
		return 0, nil
	}
	return incrementWithTTLScript.Run(ctx, r.client, []string{key}, int(ttl.Seconds())).Int64()
}

// Count returns the current integer value stored at key (0 if absent). Used to
// peek at counters (e.g. failed-login attempts) without incrementing them.
func (r *Redis) Count(ctx context.Context, key string) int64 {
	if r == nil {
		return 0
	}
	n, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		return 0
	}
	return n
}

func (r *Redis) Close() {
	if r == nil {
		return
	}
	r.client.Close()
}
