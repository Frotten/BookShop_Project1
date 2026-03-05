package redis

import (
	"Project1_Shop/settings"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init(cfg *settings.RedisConfig) (err error) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	return nil
}

func Close() {
	_ = RDB.Close()
}
