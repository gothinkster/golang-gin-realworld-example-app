package users

import "golang-gin-starter-kit/common"

type ProfileSerializer struct {
    UserModel
    myUserModel UserModel
    Following   bool
}

type ProfileResponse struct {
    ID        uint        `json:"-"`
    Username  string      `json:"username"`
    Bio       string      `json:"bio"`
    Image     *string     `json:"image"`
    Following bool        `json:"following"`
}

func (self *ProfileSerializer) Response() ProfileResponse {
    profile := ProfileResponse{
        ID:        self.ID,
        Username:  self.Username,
        Bio:       self.Bio,
        Image:     self.Image,
        Following: self.Following,
    }
    return profile
}

type UserSerializer struct {
    UserModel
}

type UserResponse struct {
    Username string      `json:"username"`
    Email    string      `json:"email"`
    Bio      string      `json:"bio"`
    Image    *string     `json:"image"`
    Token    string      `json:"token"`
}

func (self *UserSerializer) Response() UserResponse {
    user := UserResponse{
        Username: self.Username,
        Email:    self.Email,
        Bio:      self.Bio,
        Image:    self.Image,
        Token:    common.GenToken(self.ID),
    }
    return user
}
