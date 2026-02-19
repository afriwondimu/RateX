package storage

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

// RedisStorage implements Redis-backed storage
type RedisStorage struct {
    client *redis.Client
    ctx    context.Context
}

// NewRedisStorage creates a new Redis storage connection
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    return &RedisStorage{
        client: client,
        ctx:    ctx,
    }, nil
}

// Get retrieves a value from Redis
func (r *RedisStorage) Get(key string) (interface{}, bool) {
    val, err := r.client.Get(r.ctx, key).Result()
    if err == redis.Nil {
        return nil, false
    }
    if err != nil {
        return nil, false
    }
    
    var result interface{}
    if err := json.Unmarshal([]byte(val), &result); err != nil {
        return val, true
    }
    return result, true
}

// Set stores a value with TTL in Redis
func (r *RedisStorage) Set(key string, value interface{}, ttl time.Duration) {
    var data []byte
    var err error
    
    switch v := value.(type) {
    case string:
        data = []byte(v)
    default:
        data, err = json.Marshal(value)
        if err != nil {
            return
        }
    }
    
    r.client.Set(r.ctx, key, data, ttl)
}

// Delete removes a key from Redis
func (r *RedisStorage) Delete(key string) {
    r.client.Del(r.ctx, key)
}

// Clear removes all keys (use with caution in Redis!)
func (r *RedisStorage) Clear() error {
    return r.client.FlushDB(r.ctx).Err()
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
    return r.client.Close()
}

// Token Bucket Redis script for atomic operations
const tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local bucket = redis.call('hgetall', key)
local last_tokens = capacity
local last_refill = now

if #bucket > 0 then
    last_tokens = tonumber(bucket[2])
    last_refill = tonumber(bucket[4])
end

local elapsed = now - last_refill
local tokens_to_add = elapsed * rate
local current_tokens = math.min(capacity, last_tokens + tokens_to_add)

if current_tokens >= requested then
    local new_tokens = current_tokens - requested
    redis.call('hmset', key, 'tokens', new_tokens, 'last_refill', now)
    redis.call('expire', key, math.ceil(capacity/rate) + 1)
    return {1, new_tokens, capacity, now + math.ceil((capacity - new_tokens) / rate)}
else
    return {0, current_tokens, capacity, now + math.ceil(requested / rate)}
end
`

// TokenBucketAllow performs token bucket rate limiting in Redis
func (r *RedisStorage) TokenBucketAllow(key string, rate float64, capacity, n int64) (bool, *RateLimitInfo, error) {
    now := float64(time.Now().Unix())
    
    result, err := r.client.Eval(r.ctx, tokenBucketScript, []string{key}, 
        rate, capacity, now, n).Result()
    
    if err != nil {
        return false, nil, err
    }

    values := result.([]interface{})
    allowed := values[0].(int64) == 1
    remaining := int64(values[1].(float64))
    reset := time.Unix(int64(values[3].(float64)), 0)

    return allowed, &RateLimitInfo{
        Limit:     capacity,
        Remaining: remaining,
        Reset:     reset,
    }, nil
}