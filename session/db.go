package session

import (
	"github.com/jinzhu/gorm"
)

type SetdbFunc func(db *gorm.DB) error

var setdbFuncs []SetdbFunc

func RegisterSetdb(fn SetdbFunc) {
	setdbFuncs = append(setdbFuncs, fn)
}

func Setdb(s *Session) error {
	db := s.MysqlWrite().Debug()
	for _, fn := range setdbFuncs {
		if err := fn(db); err != nil {
			db.AddError(err)
		}
	}

	return db.Error
}
