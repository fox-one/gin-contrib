package session

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var ErrNotFound = errors.New("session: nil")

type Session struct {
	// db
	redis      *redis.Client
	mysqlRead  *gorm.DB
	mysqlWrite *gorm.DB

	aws *session.Session

	// shared configuration
	v *viper.Viper

	// context
	ctx context.Context
}

func New(data []byte) (*Session, error) {
	return NewWithReader(bytes.NewReader(data))
}

func NewWithReader(r io.Reader) (*Session, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}

	return NewWithViper(v), nil
}

func NewWithViper(v *viper.Viper) *Session {
	s := &Session{v: v}
	s.redis = openRedis(v)
	s.mysqlRead, s.mysqlWrite = openMysql(v)
	s.aws = awsSession(v)
	return s
}

func (s *Session) Copy() *Session {
	return &Session{
		redis:      s.redis,
		mysqlRead:  s.mysqlRead,
		mysqlWrite: s.mysqlWrite,
		aws:        s.aws,
		v:          s.v,
		ctx:        s.ctx,
	}
}

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

func (s *Session) Redis() *redis.Client {
	return s.redis
}

func (s *Session) AWSSession() *session.Session {
	return s.aws
}

func (s *Session) Viper() *viper.Viper {
	return s.v
}

func (s *Session) SubViper(key string) *Session {
	cp := s.Copy()
	cp.v = s.v.Sub(key)
	return cp
}

func (s *Session) UnmarshalViperWithTag(r interface{}, tag string) error {
	return s.Viper().Unmarshal(r, func(opt *mapstructure.DecoderConfig) {
		opt.TagName = tag
	})
}

func (s *Session) UnmarshalViper(r interface{}) error {
	return s.UnmarshalViperWithTag(r, "json")
}

func (s *Session) MysqlRead() *gorm.DB {
	return s.mysqlRead
}

func (s *Session) MysqlWrite() *gorm.DB {
	return s.mysqlWrite
}

func (s *Session) MysqlBegin() *Session {
	cp := s.Copy()
	cp.mysqlWrite = s.MysqlWrite().Begin()
	return cp
}

func (s *Session) MysqlRollback() *gorm.DB {
	return s.mysqlWrite.Rollback()
}

type sqlTx interface {
	Rollback() error
}

func (s *Session) MysqlRollbackUnlessCommitted() *gorm.DB {
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

func (s *Session) MysqlCommit() *gorm.DB {
	return s.mysqlWrite.Commit()
}

func (s *Session) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}

	return context.Background()
}

func (s *Session) WithContext(ctx context.Context) *Session {
	if ctx == nil {
		panic("nil context")
	}

	cp := s.Copy()
	cp.ctx = ctx
	return cp
}

func IsErrNotFound(err error) bool {
	switch err {
	case redis.Nil, ErrNotFound:
		return true
	default:
		return gorm.IsRecordNotFoundError(err)
	}
}
