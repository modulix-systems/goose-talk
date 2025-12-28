package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Redis struct {
	*redis.Client
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
}

func New(url string, opts ...Option) (*Redis, error) {
	rdb := &Redis{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(rdb)
	}

	redisConfig, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis - New - redis.ParseURL: %w", err)
	}

	redisConfig.MaxActiveConns = rdb.maxPoolSize
	redisConfig.DialTimeout = rdb.connTimeout
	redisConfig.DialerRetries = rdb.connAttempts

	rdb.Client = redis.NewClient(redisConfig)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis - New - rdb.Ping: %w", err)
	}

	return rdb, nil
}
