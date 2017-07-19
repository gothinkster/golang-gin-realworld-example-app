package main

import (
	"gopkg.in/gin-gonic/gin.v1"

    "golang-gin-starter-kit/common"
    "golang-gin-starter-kit/middlewares"
    "golang-gin-starter-kit/users"
)


func main() {

	db := common.DatabaseConnection()
	defer db.Close()

    db.DB().SetMaxIdleConns(10)
    db.AutoMigrate(&users.UserModel{})


    r := gin.Default()
	r.Use(middlewares.DatabaseMiddleware(db))

    usersGroup := r.Group("/api/v1/users")
    users.Register(usersGroup)

    testAuth := r.Group("/api/v1/ping")
    testAuth.Use(middlewares.Auth(common.NBSecretPassword))

    testAuth.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

	r.Run() // listen and serve on 0.0.0.0:8080
}