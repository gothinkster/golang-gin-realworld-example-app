package users

type RegistrationValidator struct {
    User struct {
        Username      string      `form:"username" json:"username" binding:"exists,alphanum,min=8,max=255"`
        Email         string      `form:"email" json:"email" binding:"exists,email"`
        Password      string      `form:"password" json:"password" binding:"exists,min=8,max=255"`
    } `json:"user"`
}

type LoginValidator struct {
    User struct {
        Email         string      `form:"email" json:"email" binding:"exists,email"`
        Password      string      `form:"password"json:"password" binding:"exists,min=8,max=255"`
    } `json:"user"`
}
