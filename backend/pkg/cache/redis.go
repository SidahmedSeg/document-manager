package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/SidahmedSeg/document-manager/backend/pkg/config"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
	"go.uber.org/zap"
)

// Cache wraps Redis client with helper methods
type Cache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(cfg config.RedisConfig, logger *zap.Logger) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(errors.ErrCodeCache, "failed to connect to Redis", err)
	}

	if logger != nil {
		logger.Info("redis connection established",
			zap.String("addr", cfg.GetRedisAddr()),
			zap.Int("db", cfg.DB),
		)
	}

	return &Cache{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	if c.logger != nil {
		c.logger.Info("closing redis connection")
	}
	return c.client.Close()
}

// Client returns the underlying Redis client
func (c *Cache) Client() *redis.Client {
	return c.client
}

// HealthCheck performs a health check on Redis
func (c *Cache) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "redis health check failed", err)
	}

	return nil
}

// Set stores a value with TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to marshal value", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to set cache",
				zap.String("key", key),
				zap.Error(err),
			)
		}
		return errors.Wrap(errors.ErrCodeCache, "failed to set cache value", err)
	}

	return nil
}

// Get retrieves a value and unmarshals it into dest
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.ErrNotFound
		}
		if c.logger != nil {
			c.logger.Error("failed to get cache",
				zap.String("key", key),
				zap.Error(err),
			)
		}
		return errors.Wrap(errors.ErrCodeCache, "failed to get cache value", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to unmarshal value", err)
	}

	return nil
}

// GetString retrieves a string value
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.ErrNotFound
		}
		return "", errors.Wrap(errors.ErrCodeCache, "failed to get string value", err)
	}
	return val, nil
}

// SetString stores a string value
func (c *Cache) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to set string value", err)
	}
	return nil
}

// Delete removes a key
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to delete cache",
				zap.Strings("keys", keys),
				zap.Error(err),
			)
		}
		return errors.Wrap(errors.ErrCodeCache, "failed to delete cache keys", err)
	}
	return nil
}

// Exists checks if a key exists
func (c *Cache) Exists(ctx context.Context, keys ...string) (bool, error) {
	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeCache, "failed to check key existence", err)
	}
	return count > 0, nil
}

// Expire sets a TTL on a key
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to set expiration", err)
	}
	return nil
}

// TTL returns the remaining TTL of a key
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeCache, "failed to get TTL", err)
	}
	return ttl, nil
}

// Incr increments a counter
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeCache, "failed to increment counter", err)
	}
	return val, nil
}

// Decr decrements a counter
func (c *Cache) Decr(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeCache, "failed to decrement counter", err)
	}
	return val, nil
}

// IncrBy increments a counter by a specific amount
func (c *Cache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeCache, "failed to increment counter", err)
	}
	return val, nil
}

// HSet sets a hash field
func (c *Cache) HSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to marshal value", err)
	}

	if err := c.client.HSet(ctx, key, field, data).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to set hash field", err)
	}
	return nil
}

// HGet retrieves a hash field
func (c *Cache) HGet(ctx context.Context, key, field string, dest interface{}) error {
	data, err := c.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.ErrNotFound
		}
		return errors.Wrap(errors.ErrCodeCache, "failed to get hash field", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to unmarshal value", err)
	}
	return nil
}

// HGetAll retrieves all hash fields
func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeCache, "failed to get hash", err)
	}
	return data, nil
}

// HDel deletes hash fields
func (c *Cache) HDel(ctx context.Context, key string, fields ...string) error {
	if err := c.client.HDel(ctx, key, fields...).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to delete hash fields", err)
	}
	return nil
}

// SAdd adds members to a set
func (c *Cache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if err := c.client.SAdd(ctx, key, members...).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to add to set", err)
	}
	return nil
}

// SMembers returns all members of a set
func (c *Cache) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeCache, "failed to get set members", err)
	}
	return members, nil
}

// SRem removes members from a set
func (c *Cache) SRem(ctx context.Context, key string, members ...interface{}) error {
	if err := c.client.SRem(ctx, key, members...).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to remove from set", err)
	}
	return nil
}

// SIsMember checks if a value is a member of a set
func (c *Cache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	isMember, err := c.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeCache, "failed to check set membership", err)
	}
	return isMember, nil
}

// FlushDB flushes the current database (use with caution!)
func (c *Cache) FlushDB(ctx context.Context) error {
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return errors.Wrap(errors.ErrCodeCache, "failed to flush database", err)
	}
	if c.logger != nil {
		c.logger.Warn("redis database flushed")
	}
	return nil
}

// BuildKey builds a cache key with prefix
func BuildKey(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	key := parts[0]
	for i := 1; i < len(parts); i++ {
		key = fmt.Sprintf("%s:%s", key, parts[i])
	}
	return key
}

// TenantKey builds a tenant-scoped cache key
func TenantKey(tenantID string, parts ...string) string {
	allParts := append([]string{"tenant", tenantID}, parts...)
	return BuildKey(allParts...)
}

// UserKey builds a user-scoped cache key
func UserKey(userID string, parts ...string) string {
	allParts := append([]string{"user", userID}, parts...)
	return BuildKey(allParts...)
}
