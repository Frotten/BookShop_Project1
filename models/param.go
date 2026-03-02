package models

type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`                     //json表示前端传过来的字段名，binding表示校验规则
	Password   string `json:"password" binding:"required"`                     //binding:"required"表示该字段为必填项
	RePassword string `json:"re_Password" binding:"required,eqfield=Password"` //eqfield=Password表示该字段必须和Password字段相等
	Email      string `json:"email" binding:"required"`                        //email表示该字段必须是一个合法的邮箱格式
	Gender     int8   `json:"gender" binding:"required,oneof=0 1"`             //oneof=0 1表示只能是这两个值之一，0表示男，1表示女
}

type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
