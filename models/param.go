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
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type AddBookParam struct {
	Title        string   `json:"title" binding:"required"`
	Author       string   `json:"author" binding:"required"`
	Publisher    string   `json:"publisher" binding:"required"`
	Introduction string   `json:"introduction"`
	Stock        int64    `json:"stock" binding:"required"`
	Price        int64    `json:"price" binding:"required"`
	CoverImage   string   `json:"cover_image"`
	Tags         []string `json:"tags"`
}

type UpdateBookParam struct {
	BookID       int64    `json:"book_id"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Publisher    string   `json:"publisher"`
	Introduction string   `json:"introduction"`
	Stock        int64    `json:"stock"`
	Price        int64    `json:"price"` //Price和Score都用int64，前端显示时除以100,避免精度误差
	CoverImage   string   `json:"cover_image" `
	Tags         []string `json:"tags" ` //切片用法
}
