package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
)

// *ModelValidator containing two parts:
// - Validator: write the form/json checking rule according to the doc https://github.com/go-playground/validator
// - DataModel: fill with data from Validator after invoking common.Bind(c, self)
// Then, you can just call model.save() after the data is ready in DataModel.
type UserModelValidator struct {
	User struct {
		Username string `form:"username" json:"username" binding:"required,min=4,max=255"`
		Email    string `form:"email" json:"email" binding:"required,email"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
		Bio      string `form:"bio" json:"bio" binding:"max=1024"`
		Image    string `form:"image" json:"image" binding:"omitempty,url"`
	} `json:"user"`
	userModel UserModel `json:"-"`
}

// You can put your general binding logic here such as setting password.
func (self *UserModelValidator) Bind(c *gin.Context) error {
	err := common.Bind(c, self)
	if err != nil {
		return err
	}
	self.userModel.Username = self.User.Username
	self.userModel.Email = self.User.Email
	self.userModel.Bio = self.User.Bio
	if err := self.userModel.setPassword(self.User.Password); err != nil {
		return err
	}
	if self.User.Image != "" {
		self.userModel.Image = &self.User.Image
	}
	return nil
}

// You can put the default value of a Validator here
func NewUserModelValidator() UserModelValidator {
	userModelValidator := UserModelValidator{}
	return userModelValidator
}

// UserUpdateValidator is the schema for PUT /api/user. Nullable fields give
// tri-state semantics (absent / null / value); the binding tags apply to the
// inner value via the valuer registered in common: "omitnil" skips absent
// fields, null and "" fail "required" on identity fields, and null on the
// tag-free bio/image fields is accepted (it clears them).
type UserUpdateValidator struct {
	User struct {
		Username common.Nullable[string] `json:"username" binding:"omitnil,required,min=4,max=255"`
		Email    common.Nullable[string] `json:"email" binding:"omitnil,required,email"`
		Password common.Nullable[string] `json:"password" binding:"omitnil,required,min=8,max=255"`
		Bio      common.Nullable[string] `json:"bio" binding:"omitnil"`
		Image    common.Nullable[string] `json:"image" binding:"omitnil"`
	} `json:"user"`
}

func NewUserUpdateValidator() UserUpdateValidator {
	return UserUpdateValidator{}
}

func (self *UserUpdateValidator) Bind(c *gin.Context) error {
	return common.Bind(c, self)
}

// invalidFields lists request fields whose JSON value had the wrong type.
// The binding tags cannot see this case: the valuer collapses it to the zero
// value, which reads as blank (identity fields) or is accepted outright
// (tag-free bio/image).
func (self *UserUpdateValidator) invalidFields() []string {
	fields := []string{}
	if self.User.Username.IsInvalid() {
		fields = append(fields, "username")
	}
	if self.User.Email.IsInvalid() {
		fields = append(fields, "email")
	}
	if self.User.Password.IsInvalid() {
		fields = append(fields, "password")
	}
	if self.User.Bio.IsInvalid() {
		fields = append(fields, "bio")
	}
	if self.User.Image.IsInvalid() {
		fields = append(fields, "image")
	}
	return fields
}

type LoginValidator struct {
	User struct {
		Email    string `form:"email" json:"email" binding:"required,email"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	} `json:"user"`
	userModel UserModel `json:"-"`
}

func (self *LoginValidator) Bind(c *gin.Context) error {
	err := common.Bind(c, self)
	if err != nil {
		return err
	}

	self.userModel.Email = self.User.Email
	return nil
}

// You can put the default value of a Validator here
func NewLoginValidator() LoginValidator {
	loginValidator := LoginValidator{}
	return loginValidator
}
