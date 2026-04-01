package Worker

import (
	"Project1_Shop/logic"
	"Project1_Shop/models"
	"context"
	"fmt"

	"go.uber.org/zap"
)

func StartRateWorker(ctx context.Context) { //设置工作池
	for i := 0; i < models.Workers; i++ {
		go rateWorker(ctx, i)
	}
}

// rateWorker 用来处理评分，不涉及金钱以及严格一致性，还拥有极低延迟，实现简单，可以作为对MQ的简易模拟
func rateWorker(ctx context.Context, id int) {
	for {
		select {
		case t, ok := <-models.RateChan:
			if !ok {
				zap.L().Info("RateChan closed, worker exiting", zap.Int("id", id))
				return
			}
			var err error
			switch t.Op {
			case models.RateOpNew:
				err = logic.NewScoreAndRank(t.UserID, t.BookID, t.Score)
			case models.RateOpUpdate:
				err = logic.UpdateScoreAndRank(t.UserID, t.BookID, t.Score)
			default:
				err = fmt.Errorf("unknown Op")
			}
			if err != nil {
				zap.L().Error("StartRateWorker failed", zap.Error(err))
			}
		case <-ctx.Done():
			zap.L().Info("worker exiting due to context cancellation", zap.Int("id", id))
			return
		}
	}
}
