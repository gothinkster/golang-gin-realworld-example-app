package users

type UserModelValidator struct {
    User struct {
        Username      string      `form:"username" json:"username" binding:"exists,alphanum,min=8,max=255"`
        Email         string      `form:"email" json:"email" binding:"exists,email"`
        Password      string      `form:"password" json:"password" binding:"exists,min=8,max=255"`
        Bio           string      `form:"bio" json:"bio" binding:"max=1024"`
        Image         string      `form:"image" json:"image" binding:"omitempty,url"`
    } `json:"user"`
}

type LoginValidator struct {
    User struct {
        Email         string      `form:"email" json:"email" binding:"exists,email"`
        Password      string      `form:"password"json:"password" binding:"exists,min=8,max=255"`
    } `json:"user"`
}
