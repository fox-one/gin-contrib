package session

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

// ErrNotFound err not found
var ErrNotFound = errors.New("session: nil")

// Session session
type Session struct {
	// db
	memory     *cache.Cache
	redis      *redis.Client
	mysqlRead  *gorm.DB
	mysqlWrite *gorm.DB

	aws *session.Session

	// shared configuration
	v *viper.Viper

	// context
	ctx context.Context
}

// New new session
func New(data []byte) (*Session, error) {
	return NewWithReader(bytes.NewReader(data))
}

// NewWithReader new session with reader
func NewWithReader(r io.Reader) (*Session, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}

	return NewWithViper(v), nil
}

// NewWithViper new session with viper
func NewWithViper(v *viper.Viper) *Session {
	s := &Session{v: v}
	s.memory = cache.New(time.Hour, time.Minute*10)
	s.redis = openRedis(v)
	s.mysqlRead, s.mysqlWrite = openMysql(v)
	s.aws = awsSession(v)
	return s
}

// Copy copy
func (s *Session) Copy() *Session {
	return &Session{
		memory:     s.memory,
		redis:      s.redis,
		mysqlRead:  s.mysqlRead,
		mysqlWrite: s.mysqlWrite,
		aws:        s.aws,
		v:          s.v,
		ctx:        s.ctx,
	}
}

// Close close
func (s *Session) Close() {
	if c := s.redis; c != nil {
		c.Close()
	}

	if c := s.mysqlRead; c != nil {
		c.Close()
	}

	if c := s.mysqlWrite; c != nil {
		c.Close()
	}
}

// Redis redis
func (s *Session) Redis() *redis.Client {
	return s.redis
}

// AWSSession aws session
func (s *Session) AWSSession() *session.Session {
	return s.aws
}

// Viper viper
func (s *Session) Viper() *viper.Viper {
	return s.v
}

// SubViper sub viper
func (s *Session) SubViper(key string) *Session {
	cp := s.Copy()
	cp.v = s.v.Sub(key)
	return cp
}

// UnmarshalViperWithTag unmarshal with tag
func (s *Session) UnmarshalViperWithTag(r interface{}, tag string) error {
	return s.Viper().Unmarshal(r, func(opt *mapstructure.DecoderConfig) {
		opt.TagName = tag
	})
}

// UnmarshalViper unmarshal
func (s *Session) UnmarshalViper(r interface{}) error {
	return s.UnmarshalViperWithTag(r, "json")
}

// MemoryCache memory cache
func (s *Session) MemoryCache() *cache.Cache {
	return s.memory
}

// MysqlRead mysql read
func (s *Session) MysqlRead() *gorm.DB {
	return s.mysqlRead
}

// MysqlWrite mysql write
func (s *Session) MysqlWrite() *gorm.DB {
	return s.mysqlWrite
}

// MysqlReadOnWrite mysql all write
func (s *Session) MysqlReadOnWrite() *Session {
	s = s.Copy()
	s.mysqlRead = s.mysqlWrite
	return s
}

// MysqlBegin mysql begin
func (s *Session) MysqlBegin() *Session {
	cp := s.Copy()
	cp.mysqlWrite = s.MysqlWrite().Begin()
	return cp
}

// MysqlRollback mysql rollback
func (s *Session) MysqlRollback() *gorm.DB {
	return s.mysqlWrite.Rollback()
}

// MysqlRollbackUnlessCommitted rollback unless committed
func (s *Session) MysqlRollbackUnlessCommitted() *gorm.DB {
	type sqlTx interface {
		Rollback() error
	}

	var emptySQLTx *sql.Tx
	if db, ok := s.mysqlWrite.CommonDB().(sqlTx); ok && db != nil && db != emptySQLTx {
		err := db.Rollback()
		// Ignore the error indicating that the transaction has already
		// been committed.
		if err != sql.ErrTxDone {
			s.mysqlWrite.AddError(err)
		}
	} else {
		s.mysqlWrite.AddError(gorm.ErrInvalidTransaction)
	}

	return s.mysqlWrite
}

// MysqlCommit mysql commit
func (s *Session) MysqlCommit() *gorm.DB {
	return s.mysqlWrite.Commit()
}

// context.Context

// Deadline deadline
func (s *Session) Deadline() (deadline time.Time, ok bool) {
	return s.Context().Deadline()
}

// Done done
func (s *Session) Done() <-chan struct{} {
	return s.Context().Done()
}

// Err err
func (s *Session) Err() error {
	return s.Context().Err()
}

// Value value
func (s *Session) Value(key interface{}) interface{} {
	return s.Context().Value(key)
}

// Context context
func (s *Session) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}

	return context.Background()
}

// WithContext with context
func (s *Session) WithContext(ctx context.Context) *Session {
	if ctx == nil {
		panic("nil context")
	}

	cp := s.Copy()
	cp.ctx = ctx
	return cp
}

// IsErrNotFound is err not found
func IsErrNotFound(err error) bool {
	switch err {
	case redis.Nil, ErrNotFound:
		return true
	default:
		return gorm.IsRecordNotFoundError(err)
	}
}

// IsErrNotFound is err not found
func (s *Session) IsErrNotFound(err error) bool {
	return IsErrNotFound(err)
}
