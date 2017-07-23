package middlewares


import (
    "testing"
    "net/http"
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/stretchr/testify/assert"
    _ "golang-gin-starter-kit/common"
    "net/http/httptest"
    "golang-gin-starter-kit/common"
)

func TestAuth(t *testing.T) {
    assert := assert.New(t)

    r := gin.New()
    v1 := r.Group("/v1")
    const pong string = "pong"
    v1.Use(Auth())
    {
        v1.GET("/ping", func(c *gin.Context) {
            my_user_id,ok := c.Get("my_user_id")
            assert.True(ok,"anonymous should work correct")
            assert.Equal(uint64(0),my_user_id,"anonymous should return user id 0")
            c.String(http.StatusOK, pong)
        })
    }

    req, err := http.NewRequest("GET", "/v1/ping", nil)
    assert.NoError(err)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal("", w.Body.String(), "response content should be nil")
    assert.Equal(http.StatusUnauthorized,w.Code,"response status should be 401")
}

func TestUnAuth(t *testing.T) {
    assert := assert.New(t)

    r := gin.New()
    v1 := r.Group("/v1")
    const pong string = "pong"
    v1.Use(Auth())
    {
        v1.GET("/ping", func(c *gin.Context) {
            my_user_id,ok := c.Get("my_user_id")
            assert.True(ok,"login user should work correct")
            assert.Equal(uint64(2),my_user_id,"login user should return user id 2")
            c.String(http.StatusOK, pong)
        })
    }

    token, err := common.GenToken(uint64(2))
    req, err := http.NewRequest("GET", "/v1/ping", nil)
    assert.NoError(err)
    req.Header.Set("Authorization", "Token "+token)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(pong, w.Body.String(), "response content should be 'pong'")
    assert.Equal(http.StatusOK, w.Code,"response status should be 200")
}