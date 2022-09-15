package limiter

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	uuid "github.com/gofrs/uuid"
)

type Blocker interface {
	BlockUntil(id, cause string, exp time.Time) error
	State(id string) (exp time.Time, cause string, blocked bool)
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

func (b *sortedSetBlocker) BlockUntil(id, cause string, exp time.Time) error {
	score := exp.Unix()
	if score <= time.Now().Unix() {
		return nil
	}

	key := b.key(id)
	_, err := b.client.Pipelined(context.Background(), func(p redis.Pipeliner) error {
		p.ZRemRangeByScore(context.Background(), key, "-inf", strconv.FormatInt(score, 10))
		p.ZAdd(context.Background(), key, redis.Z{
			Member: uuid.Must(uuid.NewV4()).String() + ":" + cause,
			Score:  float64(score),
		})
		p.Expire(context.Background(), key, b.maxAge)
		return nil
	})
	return err
}

func (b *sortedSetBlocker) State(id string) (exp time.Time, cause string, blocked bool) {
	key := b.key(id)
	if val := b.client.ZRangeWithScores(context.Background(), key, -1, -1).Val(); len(val) >= 1 {
		z := val[0]
		e := time.Unix(int64(z.Score), 0)
		if blocked = e.After(time.Now()); blocked {
			exp = e
			if member, ok := z.Member.(string); ok {
				if segments := strings.SplitN(member, ":", 2); len(segments) == 2 {
					cause = segments[1]
				}
			}
		}
	}

	return
}

func (b *sortedSetBlocker) Clean(id string) error {
	return b.client.Del(context.Background(), b.key(id)).Err()
}
