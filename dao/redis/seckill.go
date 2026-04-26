package redis

import (
	"Project1_Shop/models"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func seckillUserKey(userID, productID int64) string {
	return "seckill:user:" + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(productID, 10)
}

const seckillQueueKey = "seckill:queue"
const seckillActiveKey = "seckill:active"

var seckillLua = goredis.NewScript(`
local dup = redis.call('SET', KEYS[1], '1', 'NX', 'EX', tonumber(ARGV[1]))
if not dup then
    return 1
end
local stock = tonumber(redis.call('GET', KEYS[2]))
if stock == nil or stock <= 0 then
    redis.call('DEL', KEYS[1])
    return 2
end
redis.call('DECR', KEYS[2])
return 0
`)

func TrySeckill(userID, productID int64, dedupTTL time.Duration) (int64, error) {
	result, err := seckillLua.Run(
		ctx,
		RDB,
		[]string{
			seckillUserKey(userID, productID),
			"seckill:stock:" + strconv.FormatInt(productID, 10),
		},
		int64(dedupTTL.Seconds()),
	).Int64()
	if err != nil {
		return -1, err
	}
	return result, nil
}

func InitSeckillStock(productID, stock int64, endTime time.Time) error {
	key := "seckill:stock:" + strconv.FormatInt(productID, 10)
	ttl := time.Until(endTime) + time.Minute*10 // 多留 10 分钟保险
	if ttl <= 0 {
		ttl = time.Minute * 10
	}
	return RDB.Set(ctx, key, stock, ttl).Err()
}

func GetSeckillStock(productID int64) (int64, error) {
	key := "seckill:stock:" + strconv.FormatInt(productID, 10)
	val, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func RecoverSeckillStock(productID int64) error {
	key := "seckill:stock:" + strconv.FormatInt(productID, 10)
	return RDB.Incr(ctx, key).Err()
}

func CacheSeckillProduct(sp *models.SeckillProduct) error {
	key := "seckill:product:" + strconv.FormatInt(sp.ID, 10)
	pipe := RDB.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"id":            sp.ID,
		"book_id":       sp.BookID,
		"title":         sp.Title,
		"seckill_price": sp.SeckillPrice,
		"orig_price":    sp.OrigPrice,
		"stock":         sp.Stock,
		"start_time":    sp.StartTime.Format(models.TimeParseLayout),
		"end_time":      sp.EndTime.Format(models.TimeParseLayout),
		"status":        sp.Status,
	})
	ttl := time.Until(sp.EndTime) + time.Minute*10
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	pipe.ZAdd(ctx, seckillActiveKey, goredis.Z{
		Score:  float64(sp.EndTime.Unix()),
		Member: sp.ID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func GetSeckillProductFromCache(productID int64) (*models.SeckillProductView, error) {
	key := "seckill:product:" + strconv.FormatInt(productID, 10)
	data, err := RDB.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, err
	}
	id, _ := strconv.ParseInt(data["id"], 10, 64)
	bookID, _ := strconv.ParseInt(data["book_id"], 10, 64)
	sp, _ := strconv.ParseInt(data["seckill_price"], 10, 64)
	op, _ := strconv.ParseInt(data["orig_price"], 10, 64)
	status, _ := strconv.ParseInt(data["status"], 10, 64)
	stock, _ := GetSeckillStock(productID)
	return &models.SeckillProductView{
		ID:           id,
		BookID:       bookID,
		Title:        data["title"],
		SeckillPrice: sp,
		OrigPrice:    op,
		Stock:        stock,
		StartTime:    data["start_time"],
		EndTime:      data["end_time"],
		Status:       int8(status),
	}, nil
}

func GetActiveSeckillIDs() ([]int64, error) {
	now := float64(time.Now().Unix())
	members, err := RDB.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     seckillActiveKey,
		Start:   now,
		Stop:    -1,
		ByScore: true,
		Rev:     false,
	}).Result()
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, m := range members {
		id, _ := strconv.ParseInt(m, 10, 64)
		ids = append(ids, id)
	}
	return ids, nil
}

func RemoveSeckillFromActive(productID int64) error {
	pipe := RDB.Pipeline()
	pipe.ZRem(ctx, seckillActiveKey, productID)
	SPK := "seckill:product:" + strconv.FormatInt(productID, 10)
	SSK := "seckill:stock:" + strconv.FormatInt(productID, 10)
	pipe.Del(ctx, SPK)
	pipe.Del(ctx, SSK)
	_, err := pipe.Exec(ctx)
	return err
}

func RevokeSeckillUser(userID, productID int64) error {
	return RDB.Del(ctx, seckillUserKey(userID, productID)).Err()
}
