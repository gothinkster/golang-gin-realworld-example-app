package main

import (
	"gopkg.in/gin-gonic/gin.v1"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

    "golang-gin-starter-kit/storage"
    "golang-gin-starter-kit/users"
)


func main() {

	db := storage.DatabaseConnection()
	defer db.Close()

    db.AutoMigrate(&users.UserModel{})


    r := gin.Default()
	r.Use(storage.ApiMiddleware(db))

    users.Register(r.Group("/api/v1/users"))

	r.Run() // listen and serve on 0.0.0.0:8080
}