package middlewares

import (
    "testing"
    "github.com/jinzhu/gorm"
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/stretchr/testify/assert"
    "golang-gin-starter-kit/common"
    "net/http"
    "net/http/httptest"
)

func TestDatabaseMiddleware(t *testing.T) {
    assert := assert.New(t)

    dbConnection := common.DatabaseConnection()
    dbConnection2 := common.DatabaseConnection()
    defer dbConnection.Close()
    defer dbConnection2.Close()
    r := gin.New()
    v1 := r.Group("/v1")
    const pong string = "pong"
    v1.Use(DatabaseMiddleware(dbConnection))
    {
        v1.GET("/ping", func(c *gin.Context) {
            db := c.MustGet("DB").(*gorm.DB)
            assert.Equal(dbConnection, db,"db in middleware should be same with outer one")
            assert.NotEqual(dbConnection2, db,"db in middleware should not be same as another")

            assert.NoError(db.DB().Ping(),"Db should be able to ping")
            c.String(http.StatusOK, pong)
        })
    }

    req, err := http.NewRequest("GET", "/v1/ping", nil)
    assert.NoError(err)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(pong, w.Body.String(), "response content should be 'pong'")
    assert.Equal(http.StatusOK, w.Code,"response status should be 200")
}
