package session

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

func dialMysql(host, user, pwd, database string) *gorm.DB {
	path := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=True&charset=utf8mb4",
		user,
		pwd,
		"tcp",
		host,
		database,
	)

	db, err := gorm.Open("mysql", path)
	if err != nil {
		log.Panicf("Connect mysql %s failed: %s", host, err)
	}

	db.DB().SetMaxIdleConns(10)
	return db
}

func openMysql(v *viper.Viper) (r, w *gorm.DB) {
	v = v.Sub("mysql")
	if v == nil {
		return nil, nil
	}

	var (
		username = v.GetString("username")
		password = v.GetString("password")
		db       = v.GetString("db")
		read     = v.GetString("read")
		write    = v.GetString("write")
	)

	r = dialMysql(read, username, password, db)
	if write == "" || write == read {
		w = r
	} else {
		w = dialMysql(write, username, password, db)
	}

	return
}

func openRedis(v *viper.Viper) *redis.Client {
	v = v.Sub("redis")
	if v == nil {
		return nil
	}

	var (
		addr = v.GetString("addr")
		db   = v.GetInt("db")
	)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       db,
	})

	if err := client.Ping().Err(); err != nil {
		panic(err)
	}

	return client
}

func awsSession(v *viper.Viper) *session.Session {
	v = v.Sub("aws")
	if v == nil {
		return nil
	}

	var (
		key    = v.GetString("key")
		secret = v.GetString("secret")
		region = v.GetString("region")
	)

	return session.New(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	})
}
