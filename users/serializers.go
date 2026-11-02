package users

import (
	"github.com/gin-gonic/gin"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
)

type ProfileSerializer struct {
	C *gin.Context
	UserModel
}

// Declare your response schema here
type ProfileResponse struct {
	ID        uint    `json:"-"`
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

// The RealWorld spec serializes unset bio/image as null, never as "".
func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nullableStringPtr(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

// Put your response logic including wrap the userModel here.
func (self *ProfileSerializer) Response() ProfileResponse {
	myUserModel := self.C.MustGet("my_user_model").(UserModel)
	profile := ProfileResponse{
		ID:        self.ID,
		Username:  self.Username,
		Bio:       nullableString(self.Bio),
		Image:     nullableStringPtr(self.Image),
		Following: myUserModel.isFollowing(self.UserModel),
	}
	return profile
}

type UserSerializer struct {
	c *gin.Context
}

type UserResponse struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
	Token    string  `json:"token"`
}

func (self *UserSerializer) Response() UserResponse {
	myUserModel := self.c.MustGet("my_user_model").(UserModel)
	user := UserResponse{
		Username: myUserModel.Username,
		Email:    myUserModel.Email,
		Bio:      nullableString(myUserModel.Bio),
		Image:    nullableStringPtr(myUserModel.Image),
		Token:    common.GenToken(myUserModel.ID),
	}
	return user
}
