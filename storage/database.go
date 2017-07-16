package storage

import (
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/jinzhu/gorm"
    "fmt"
)

func DatabaseConnection() *gorm.DB {
    db, err := gorm.Open("sqlite3", "gorm.db")
    if err != nil {
        fmt.Println("db err: ",err)
    }
    return db
}

func ApiMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("DB", db)
        c.Next()
    }
}
