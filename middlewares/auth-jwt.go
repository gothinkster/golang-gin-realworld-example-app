package middlewares


import (
    "net/http"
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/dgrijalva/jwt-go"
    "github.com/dgrijalva/jwt-go/request"
    "golang-gin-starter-kit/common"
)

func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token, err := request.ParseFromRequest(c.Request, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
            b := ([]byte(common.NBSecretPassword))
            return b, nil
        })
        if err != nil {
            c.AbortWithError(http.StatusUnauthorized, err)
            return
        }
        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            my_user_id := uint64(claims["id"].(float64))
            //fmt.Println(my_user_id,claims["id"])
            c.Set("my_user_id", my_user_id)
        } else {
            c.Set("my_user_id", uint64(0))
        }
    }
}
