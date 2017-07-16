package common

import (
    "math/rand"
    "time"
    "github.com/jinzhu/gorm"
    "fmt"
    
    "github.com/dgrijalva/jwt-go"

)

func DatabaseConnection() *gorm.DB {
    db, err := gorm.Open("sqlite3", "gorm.db")
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

func GenToken(id uint) (string, error){
    token := jwt.New(jwt.GetSigningMethod("HS256"))
    // Set some claims
    token.Claims = jwt.MapClaims{
        "Id":  id,
        "exp": time.Now().Add(time.Hour * 24).Unix(),
    }
    // Sign and get the complete encoded token as a string
    return token.SignedString([]byte(NBSecretPassword))
}