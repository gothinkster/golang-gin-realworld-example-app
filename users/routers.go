package users

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"golang-gin-starter-kit/common"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
    "errors"
)

type Router struct {
	BasePath string
}

func UsersRegister(router *gin.RouterGroup) Router{
	r := Router{}
	r.BasePath = router.BasePath()
	router.POST("/", r.Registration)
	router.POST("/login", r.Login)
	return r
}

func UserRegister(router *gin.RouterGroup) Router{
    r := Router{}
    r.BasePath = router.BasePath()
    router.GET("/", r.Retrieve)
    router.PUT("/", r.Update)
    return r
}


func (r Router) Registration(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)

	var validator UserModelValidator
	if err := common.Bind(c, &validator); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	var userModel UserModel
	userModel.Username = validator.User.Username
	userModel.Email = validator.User.Email
	userModel.setPassword(validator.User.Password)
	if err := db.Save(&userModel).Error; err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database",err))
		return
	}
    userModel.setToken()
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
}

func (r Router) Login(c *gin.Context) {
	db := c.MustGet("DB").(*gorm.DB)
	var validator LoginValidator
	if err := common.Bind(c, &validator); err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	var userModel UserModel

	if err := db.Where(&UserModel{Email: validator.User.Email}).First(&userModel).Error; err != nil {
        c.JSON(http.StatusForbidden, common.NewError("login",errors.New("Not Registered email or invalid password")))
		return
	}
	fmt.Println("user from DB: ", userModel)

	err := userModel.checkPassword(validator.User.Password)
	if err != nil {
        c.JSON(http.StatusForbidden, common.NewError("login",errors.New("Not Registered email or invalid password")))
		return
	}
    userModel.setToken()
	c.JSON(http.StatusOK, gin.H{"user": userModel})
}

func (r Router) Retrieve(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)
    my_user_id := c.MustGet("my_user_id")

    var userModel UserModel
    if err := db.First(&userModel,my_user_id).Error; err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewError("database",err))
        return
    }
    userModel.setToken()
    c.JSON(http.StatusCreated, gin.H{"user": userModel})
}

func (r Router) Update(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)
    my_user_id := c.MustGet("my_user_id")

    var userModel UserModel
    if err := db.First(&userModel,my_user_id).Error; err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewError("database",err))
        return
    }
    var validator UserModelValidator
    validator.User.Username = userModel.Username
    validator.User.Email = userModel.Email
    validator.User.Password = userModel.PasswordHash
    validator.User.Bio = userModel.Bio
    if userModel.Image!=nil{
        validator.User.Image = *userModel.Image
    }
    if err := common.Bind(c, &validator); err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
        return
    }
    userModel.Username = validator.User.Username
    userModel.Email = validator.User.Email
    userModel.Bio = validator.User.Bio
    userModel.Image = &validator.User.Image

    if validator.User.Image==""{
        userModel.Image = nil
    }
    if validator.User.Password!=userModel.PasswordHash{
        userModel.setPassword(validator.User.Password)
    }

    if err := db.Save(&userModel).Error; err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewError("database",err))
        return
    }
    userModel.setToken()
    c.JSON(http.StatusCreated, gin.H{"user": userModel})
}
