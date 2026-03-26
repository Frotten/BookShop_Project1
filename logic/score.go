package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"math"
	"strconv"

	"go.uber.org/zap"
)

func NewScoreAndRank(UserID, BookID, Score int64) error {
	err := redis.CacheBookScoreCount(BookID, 1)
	if err != nil {
		return err
	}
	err = redis.CacheBookScoreSum(BookID, Score)
	if err != nil {
		return err
	}
	err = redis.UpdateUserRate(UserID, BookID, Score)
	if err != nil {
		return err
	}
	AllScore, Count, err := redis.GetAllScoreAndCount(BookID)
	if err != nil {
		return err
	}
	AnsScore := models.WeightedCalculation(AllScore, Count)
	err = redis.AddScoreRank(BookID, AnsScore)
	if err != nil {
		return err
	}
	Book, err := mysql.GetBookByID(BookID)
	if err != nil {
		return err
	}
	err = redis.SetBook(Book, AnsScore*100)
	if err != nil {
		return err
	}
	return nil
}

func UpdateScoreAndRank(UserID, BookID, Score int64) error {
	BeforeScore, err := redis.GetBeforeBookScore(UserID, BookID)
	if err != nil {
		return err
	}
	err = redis.CacheBookScoreSum(BookID, Score-BeforeScore) //评分更新需要先获取用户之前的评分，然后计算新的评分差值，再更新Redis中的评分总和
	if err != nil {
		return err
	}
	err = redis.UpdateUserRate(UserID, BookID, Score)
	if err != nil {
		return err
	}
	AllScore, Count, err := redis.GetAllScoreAndCount(BookID)
	if err != nil {
		return err
	}
	AnsScore := models.WeightedCalculation(AllScore, Count)
	err = redis.AddScoreRank(BookID, AnsScore)
	if err != nil {
		return err
	}
	Book, err := mysql.GetBookByID(BookID)
	if err != nil {
		return err
	}
	err = redis.UpdateBook(Book, AnsScore*100)
	if err != nil {
		return err
	}
	return nil
}

func RateNewBook(p *models.UserRateBook) models.ResCode { //这里必须保证MySQL先成功才能去缓存Redis，不能并发
	RB, err := mysql.GetRateBookByID(p.BookID)
	if err != nil {
		return models.CodeInvalidParam
	}
	RB.ScoreCount++
	RB.Score += p.Score
	//更新MySQL数据库中的评分信息
	err = mysql.UpdateRateBook(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateUserRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateBookScore(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	//MySQL和Redis的分界线
	select {
	case models.RateChan <- p:
	default:
		go func() {
			err := NewScoreAndRank(p.UserID, p.BookID, p.Score)
			if err != nil {
				zap.L().Error("NewScoreAndUpdateRedis Failed", zap.Error(err))
			}
		}()
	}
	return models.CodeSuccess
}

func RateUpdateBook(p *models.UserRateBook) models.ResCode {
	RB, err := mysql.GetRateBookByID(p.BookID)
	BeforeScore, err := mysql.GetBeforeBookScore(p.BookID, p.UserID)
	if err != nil {
		return models.CodeServerBusy
	}
	RB.Score = RB.Score + p.Score - BeforeScore
	//更新MySQL数据库中的评分信息
	err = mysql.UpdateRateBook(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateUserRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateBookScore(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	select {
	case models.RateChan <- p:
	default:
		go func() {
			err := UpdateScoreAndRank(p.UserID, p.BookID, p.Score)
			if err != nil {
				zap.L().Error("UpdateScoreAndUpdateRedis Failed", zap.Error(err))
			}
		}()
	}
	return models.CodeSuccess
}

func RateBook(p *models.UserRateBook) models.ResCode {
	ok, err := redis.CheckRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	if !ok {
		ok = mysql.CheckRate(p)
		if ok {
			redis.UpdateUserRate(p.UserID, p.BookID, p.Score) //获取用户ID，书籍ID，上次评分
			p.Op = models.RateOpNew
			return RateUpdateBook(p)
		}
		p.Op = models.RateOpNew
		return RateNewBook(p)
	}
	p.Op = models.RateOpUpdate
	return RateUpdateBook(p)
}

func GetTopScoreList() ([]*models.ListBook, models.ResCode) {
	results, err := redis.GetScoreList()
	if err != nil { //查找失败，默认原排行榜丢失，从MySQL中重新获取排行榜
		Books, _, err := mysql.GetBooksPageByScore(1)
		if err != nil {
			return nil, models.CodeServerBusy
		}
		var AnsList []*models.ListBook
		for _, Book := range Books {
			T := BookListToCache(Book)
			_ = redis.SetBookSummary(T)
			err = redis.AddScoreRank(Book.BookID, float64(Book.Score)/100)
			if err != nil {
				continue
			}
			AnsList = append(AnsList, T)
		}
		return AnsList, models.CodeSuccess
	}
	var Ans []*models.ListBook
	for _, z := range results {
		var BookID int64
		switch v := z.Member.(type) {
		case int64:
			BookID = v
		case string:
			var err error
			BookID, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				zap.L().Error("ParseInt failed", zap.String("member", v), zap.Error(err))
				continue
			}
		default:
			zap.L().Error("unexpected member type", zap.Any("value", v))
			continue
		}
		Score := math.Trunc(z.Score*100) / 100
		T, err := redis.GetBookSummaryByBookID(BookID, Score)
		if err != nil {
			continue
		}
		if T.BookID == -1 {
			// 合并并发请求，只有一个会执行函数体
			v, err, _ := redis.G.Do(strconv.FormatInt(BookID, 10), func() (interface{}, error) {
				Book, err := mysql.GetBookByID(BookID)
				if err != nil {
					return nil, err
				}
				T := BookListToCache(Book)
				_ = redis.SetBookSummary(T)
				return T, nil
			})
			if err != nil {
				continue
			}
			T = v.(*models.ListBook)
		}
		if T.Score < 0 {
			continue
		}
		Ans = append(Ans, T)
	}
	return Ans, models.CodeSuccess
}
