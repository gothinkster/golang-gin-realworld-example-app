package users

import (
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/jinzhu/gorm"
    "net/http"
    "fmt"
    "golang-gin-starter-kit/common"
)

type Router struct {

}

func Register(router *gin.RouterGroup){
    r :=  Router{}
    router.POST("/", r.Registration)
    router.POST("/login", r.Login)
}

func (r *Router) Registration(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)
    var validator RegistrationValidator
    if err := c.Bind(&validator); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"Data binding error" : common.ErrsToList(err)})
        return
    }
    var userModel UserModel
    userModel.Username = validator.User.Username
    userModel.Email = validator.User.Email
    userModel.setPassword(validator.User.Password)
    if err := db.Save(&userModel).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"Database error" : err})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"user":userModel})
}

func (r *Router) Login(c *gin.Context) {
    db := c.MustGet("DB").(*gorm.DB)
    var validator LoginValidator
    if err := c.Bind(&validator); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"Data binding error" : common.ErrsToList(err)})
        return
    }
    var userModel UserModel

    if err := db.Where(&UserModel{ Email: validator.User.Email}).First(&userModel).Error; err != nil {
        c.JSON(http.StatusForbidden, gin.H{"Database error" : err})
        return
    }
    fmt.Println("user from DB: ", userModel)

    err := userModel.checkPassword(validator.User.Password)
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"Error" : "password error!"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"user":userModel})
}

