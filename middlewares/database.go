package middlewares

import (
    "github.com/jinzhu/gorm"
    "gopkg.in/gin-gonic/gin.v1"
)

func DatabaseMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("DB", db)
        c.Next()
    }
}
