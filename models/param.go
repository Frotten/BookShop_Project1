package models

type ParamSignUp struct {
	Username   string `json:"username" binding:"required" form:"username"`                        //json表示前端传过来的字段名，binding表示校验规则
	Password   string `json:"password" binding:"required" form:"password"`                        //binding:"required"表示该字段为必填项
	RePassword string `json:"re_Password" binding:"required,eqfield=Password" form:"re_Password"` //eqfield=Password表示该字段必须和Password字段相等
	Email      string `json:"email" binding:"required" form:"email"`                              //email表示该字段必须是一个合法的邮箱格式
	Gender     int8   `json:"gender" binding:"oneof=0 1" form:"gender"`
	//oneof=0 1表示只能是这两个值之一，1表示男，0表示女,不能有required，因为有默认值0，如果有required，就会导致gender字段无法省略，必须传入一个值，如果传入0，就会被认为是缺少gender字段，从而失败
}

type ParamLogin struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}
