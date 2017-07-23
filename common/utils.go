package common

import (
    "math/rand"
    "time"
    "github.com/jinzhu/gorm"
    "fmt"
    
    "github.com/dgrijalva/jwt-go"
    "gopkg.in/go-playground/validator.v8"
    _ "github.com/jinzhu/gorm/dialects/sqlite"

    "github.com/gin-gonic/gin/binding"
    "gopkg.in/gin-gonic/gin.v1"
)

func DatabaseConnection() *gorm.DB {
    db, err := gorm.Open("sqlite3", "./../gorm.db")
    if err != nil {
        fmt.Println("db err: ",err)
    }
    return db
}


var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

const NBSecretPassword = "heheda";

func GenToken(id uint64) (string, error){
    token := jwt.New(jwt.GetSigningMethod("HS256"))
    // Set some claims
    token.Claims = jwt.MapClaims{
        "id":  id,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    }
    // Sign and get the complete encoded token as a string
    return token.SignedString([]byte(NBSecretPassword))
}

func ErrsToList(err error) ([]interface{}){
    errs := err.(validator.ValidationErrors)
    var res []interface{}
    for _, v := range errs {
        // can translate each error one at a time.
        //fmt.Println(v.Value)
        res = append(res, v.Field)
    }
    return res
}

type CommonError struct {
    Errors map[string]interface{} `json:"errors"`
}

func NewValidatorError(err error) CommonError{
    res := CommonError{}
    res.Errors = make(map[string]interface{})
    errs := err.(validator.ValidationErrors)
    for _, v := range errs {
        // can translate each error one at a time.
        //fmt.Println("gg",v.NameNamespace)
        if v.Param !=""{
            res.Errors[v.Field]= fmt.Sprintf("{%v: %v}",v.Tag,v.Param)
        }else {
            res.Errors[v.Field]= fmt.Sprintf("{key: %v}",v.Tag)
        }

    }
    return res
}

func NewError(key string,err error) CommonError{
    res := CommonError{}
    res.Errors = make(map[string]interface{})
    res.Errors[key] = err.Error()
    return res
}

func Bind(c *gin.Context, obj interface{}) error {
    b := binding.Default(c.Request.Method, c.ContentType())
    return c.ShouldBindWith(obj, b)
}