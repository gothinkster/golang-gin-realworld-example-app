package middlewares


import (
    "testing"
    "net/http"
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/stretchr/testify/assert"
    "golang-gin-starter-kit/common"
    "net/http/httptest"
)

func TestAuth(t *testing.T) {
    assert := assert.New(t)

    r := gin.New()
    v1 := r.Group("/v1")
    const pong string = "pong"
    v1.Use(Auth(common.NBSecretPassword))
    {
        v1.GET("/ping", func(c *gin.Context) {
            c.String(http.StatusOK, pong)
        })
    }

    req, err := http.NewRequest("GET", "/v1/ping", nil)
    assert.NoError(err)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal("", w.Body.String(), "response content should be nil")
    assert.Equal(http.StatusUnauthorized,w.Code,"response status should be 401")



    token, err := common.GenToken(2)
    req, err = http.NewRequest("GET", "/v1/ping", nil)
    assert.NoError(err)
    req.Header.Set("Authorization", "Bearer "+token)
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(pong, w.Body.String(), "response content should be 'pong'")
    assert.Equal(http.StatusOK, w.Code,"response status should be 200")
}
