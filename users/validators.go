package users

type RegistrationValidator struct {
    User struct {
        Username      string      `json:"username"`
        Email         string      `json:"email"`
        Password      string      `json:"password"`
    } `json:"user"`
}

type LoginValidator struct {
    User struct {
        Email         string      `json:"email"`
        Password      string      `json:"password"`
    } `json:"user"`
}
