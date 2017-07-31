package main

import (
	"gopkg.in/gin-gonic/gin.v1"

    "golang-gin-starter-kit/common"
    "golang-gin-starter-kit/middlewares"
    "golang-gin-starter-kit/users"
    _ "fmt"
    "fmt"
    "github.com/jinzhu/gorm"
    "golang-gin-starter-kit/articles"
)

func Migrate(db *gorm.DB)  {
    db.AutoMigrate(&users.UserModel{})
    db.AutoMigrate(&users.FollowModel{})
    db.AutoMigrate(&articles.ArticleModel{})
    db.AutoMigrate(&articles.TagModel{})
}


func main() {

	db := common.Init()
    Migrate(db)
	defer db.Close()


    r := gin.Default()

    v1 := r.Group("/api")
    users.UsersRegister(v1.Group("/users"))

    v1.Use(middlewares.Auth())
    users.UserRegister(v1.Group("/user"))
    users.ProfileRegister(v1.Group("/profiles"))

    articles.ArticlesRegister(v1.Group("/articles"))


    testAuth := r.Group("/api/ping")

    testAuth.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    // test 1 to 1
    tx1 := db.Begin()
    tx1.Save(&users.UserModel{
        Username:"AAAAAAAAAAAAAAAA",
        Email:"aaaa@g.cn",
        Bio:"hehddeda",
        Image: nil,
    })
    tx1.Commit()
    var userA users.UserModel
    fmt.Println(userA)


	r.Run() // listen and serve on 0.0.0.0:8080
}