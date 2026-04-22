// 在 faultCode.go 的 const 块末尾追加以下错误码：

// CodeSeckillNotFound  = 1016  // 秒杀活动不存在或未开始
// CodeSeckillEnded     = 1017  // 秒杀活动已结束
// CodeSeckillDuplicate = 1018  // 您已参与该秒杀，请勿重复抢购
// CodeSeckillSoldOut   = 1019  // 秒杀商品已抢完

// 在 codeMsgMap 中追加：
// CodeSeckillNotFound:  "秒杀活动不存在或未开始",
// CodeSeckillEnded:     "秒杀活动已结束",
// CodeSeckillDuplicate: "您已参与该秒杀，请勿重复抢购",
// CodeSeckillSoldOut:   "秒杀商品已抢完",

// ========== 完整的 faultCode.go（供直接替换）==========
package models

type ResCode int64

type ResponseData struct {
	Code ResCode     `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy
	CodeInvalidToken
	CodeNeedLogin
	CodeBookExist
	CodeBookNotExist
	CodeListError
	CodeMySQLError
	CodeRedisError
	CodeInSufficient
	CodeOrderNotExist
	CodeOrderAlreadyConfirmed
	CodeSeckillNotFound   // 1016 秒杀活动不存在或未开始
	CodeSeckillEnded      // 1017 秒杀活动已结束
	CodeSeckillDuplicate  // 1018 已参与该秒杀
	CodeSeckillSoldOut    // 1019 秒杀商品已抢完
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:               "success",
	CodeInvalidParam:          "请求参数错误",
	CodeUserExist:             "用户已存在",
	CodeUserNotExist:          "用户不存在",
	CodeInvalidPassword:       "用户名或密码错误",
	CodeServerBusy:            "服务器繁忙，请稍后再试",
	CodeInvalidToken:          "无效的Token",
	CodeNeedLogin:             "需要登录",
	CodeBookExist:             "书籍已存在",
	CodeBookNotExist:          "书籍不存在",
	CodeListError:             "列表存在问题",
	CodeMySQLError:            "数据库错误",
	CodeRedisError:            "缓存错误",
	CodeInSufficient:          "库存不足",
	CodeOrderNotExist:         "订单不存在或无权访问",
	CodeOrderAlreadyConfirmed: "订单已确认，请勿重复操作",
	CodeSeckillNotFound:       "秒杀活动不存在或未开始",
	CodeSeckillEnded:          "秒杀活动已结束",
	CodeSeckillDuplicate:      "您已参与该秒杀，请勿重复抢购",
	CodeSeckillSoldOut:        "秒杀商品已抢完",
}

func (rc ResCode) Msg() string {
	msg, ok := codeMsgMap[rc]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
