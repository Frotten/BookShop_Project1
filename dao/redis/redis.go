package redis

import (
	"Project1_Shop/settings"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var RDB *redis.ClusterClient

var G singleflight.Group

var ctx = context.Background()

const BufferTime = time.Minute * 3
const UserTime = time.Minute * 30
const ListTime = time.Hour * 24 * 7
const CommentListTime = time.Minute * 15
const BookTime = time.Hour * 24 * 7 * 2
const OrderTime = time.Minute * 30

//func Init(cfg *settings.RedisConfig) (err error) {
//	RDB = redis.NewClient(&redis.Options{
//		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
//		Password: cfg.Password,
//		DB:       cfg.DB,
//		PoolSize: cfg.PoolSize,
//	})
//	return nil
//}

func Init(cfg *settings.RedisConfig) (err error) {
	RDB = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			cfg.Host + ":7000",
			cfg.Host + ":7001",
			cfg.Host + ":7002",
		},
		Password: cfg.Password,
		PoolSize: cfg.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

func Close() {
	_ = RDB.Close()
}
