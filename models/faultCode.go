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
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户已存在",
	CodeUserNotExist:    "用户不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeServerBusy:      "服务器繁忙，请稍后再试",
	CodeInvalidToken:    "无效的Token",
	CodeNeedLogin:       "需要登录",
	CodeBookExist:       "书籍已存在",
	CodeBookNotExist:    "书籍不存在",
	CodeListError:       "列表存在问题",
	CodeMySQLError:      "数据库错误",
	CodeRedisError:      "缓存错误",
}

func (rc ResCode) Msg() string {
	msg, ok := codeMsgMap[rc]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
