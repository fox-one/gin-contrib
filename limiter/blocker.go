package limiter

import (
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/go-redis/redis"
)

type Blocker interface {
	BlockUntil(id string, exp time.Time) error
	State(id string) (exp time.Time, blocked bool)
	Clean(id string) error
}

type sortedSetBlocker struct {
	client *redis.Client
	maxAge time.Duration
}

func NewBlocker(c *redis.Client, maxAge time.Duration) Blocker {
	return &sortedSetBlocker{c, maxAge}
}

func (b *sortedSetBlocker) key(id string) string {
	return "limiter:blocker:" + id
}

func (b *sortedSetBlocker) BlockUntil(id string, exp time.Time) error {
	score := exp.Unix()
	if score <= time.Now().Unix() {
		return nil
	}

	key := b.key(id)
	_, err := b.client.Pipelined(func(p redis.Pipeliner) error {
		p.ZRemRangeByScore(key, "-inf", strconv.FormatInt(score, 10))
		p.ZAdd(key, redis.Z{
			Member: uuid.Must(uuid.NewV4()),
			Score:  float64(score),
		})
		p.Expire(key, b.maxAge)
		return nil
	})
	return err
}

func (b *sortedSetBlocker) State(id string) (exp time.Time, blocked bool) {
	key := b.key(id)
	if val := b.client.ZRangeWithScores(key, -1, -1).Val(); len(val) >= 1 {
		z := val[0]
		exp = time.Unix(int64(z.Score), 0)
		blocked = exp.After(time.Now())
	}

	return
}

func (b *sortedSetBlocker) Clean(id string) error {
	return b.client.Del(b.key(id)).Err()
}
