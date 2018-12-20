package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

type opt struct {
	max    int
	window time.Duration
}

type Limiter struct {
	pool *redis.Client
	opts map[string]*opt
	mux  sync.Mutex
}

func NewLimiter(addr string, password string, db int) (*Limiter, error) {
	options := &redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  60 * time.Second,
		PoolSize:     512,
	}

	redisPool := redis.NewClient(options)
	return NewLimiterWithClient(redisPool)
}

func NewLimiterWithClient(c *redis.Client) (*Limiter, error) {
	if err := c.Ping().Err(); err != nil {
		return nil, err
	}

	return &Limiter{
		pool: c,
		opts: make(map[string]*opt, 0),
	}, nil
}

func (limiter *Limiter) AddGroup(group string, max int, window time.Duration) {
	limiter.mux.Lock()
	limiter.opts[group] = &opt{max, window}
	limiter.mux.Unlock()
}

func limiterKey(group, key string) string {
	return fmt.Sprintf("limiter:%s:%s", group, key)
}

func (limiter *Limiter) Available(key, group string, weight int) (int, error) {
	var (
		max    = 0
		window = time.Second
	)

	limiter.mux.Lock()
	if opt, _ := limiter.opts[group]; opt != nil {
		max, window = opt.max, opt.window
	}
	limiter.mux.Unlock()

	if max < weight {
		return max - weight, nil
	}

	now := time.Now()
	key = limiterKey(group, key)
	var zcount *redis.IntCmd
	_, err := limiter.pool.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.ZRemRangeByScore(key, "-inf", fmt.Sprint(now.Add(-window).UnixNano()/1000000))
		if weight > 0 {
			members := make([]redis.Z, 0, weight)
			score := float64(now.UnixNano() / 1000000)
			for idx := 0; idx < weight; idx += 1 {
				mem, _ := uuid.NewV4()
				members = append(members, redis.Z{Score: score, Member: mem.String()})
			}
			pipe.ZAdd(key, members...)
		}
		pipe.Expire(key, time.Second*time.Duration(int64(window.Seconds())+60))
		zcount = pipe.ZCount(key, "-inf", "+inf")
		return nil
	})
	if err != nil {
		return 0, err
	}
	count, err := zcount.Result()
	return max - int(count), err
}

func (limiter *Limiter) Clear(key, group string) error {
	key = limiterKey(group, key)
	zcount := limiter.pool.Del(key)
	_, err := zcount.Result()
	return err
}
