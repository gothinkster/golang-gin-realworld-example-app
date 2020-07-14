package main

import (
	"fmt"
	"golang-gin-realworld-example-app/articles"
	"golang-gin-realworld-example-app/common"
	"golang-gin-realworld-example-app/users"

	"github.com/jinzhu/gorm"

	// gin "gopkg.in/gin-gonic/gin.v1"
	"github.com/gin-gonic/gin"
	// "github.com/wangzitian0/golang-gin-starter-kit/articles"
	// "github.com/wangzitian0/golang-gin-starter-kit/common"
	// "github.com/wangzitian0/golang-gin-starter-kit/users"
)

func migrate(db *gorm.DB) {
	// db.DropTable(&articles.ArticleModel{}, &articles.TagModel{}, &articles.FavoriteModel{}, &articles.ArticleUserModel{}, &articles.CommentModel{}, &users.UserModel{}, &users.FollowModel{})
	users.AutoMigrate()
	db.AutoMigrate(&articles.ArticleModel{})
	db.AutoMigrate(&articles.TagModel{})
	db.AutoMigrate(&articles.FavoriteModel{})
	db.AutoMigrate(&articles.ArticleUserModel{})
	db.AutoMigrate(&articles.CommentModel{})
	db.AutoMigrate(&articles.CommentModelVote{})

}

func main() {

	db := common.Init()
	migrate(db)
	defer db.Close()

	r := gin.Default()

	v1 := r.Group("/api")
	users.UsersRegister(v1.Group("/users"))
	v1.Use(users.AuthMiddleware(false))
	articles.ArticlesAnonymousRegister(v1.Group("/articles"))
	articles.TagsAnonymousRegister(v1.Group("/tags"))

	v1.Use(users.AuthMiddleware(true))
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
	userA := users.UserModel{
		Username: "AAAAAAAAAAAAAAAA",
		Email:    "aaaa@g.cn",
		Bio:      "hehddeda",
		Image:    nil,
	}
	tx1.Save(&userA)
	tx1.Commit()
	fmt.Println(userA)

	//db.Save(&ArticleUserModel{
	//    UserModelID:userA.ID,
	//})
	//var userAA ArticleUserModel
	//db.Where(&ArticleUserModel{
	//    UserModelID:userA.ID,
	//}).First(&userAA)
	//fmt.Println(userAA)

	r.Run() // listen and serve on 0.0.0.0:8080
}
