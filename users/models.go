package users

import (
    "golang.org/x/crypto/bcrypt"
    "golang-gin-starter-kit/common"
)

type UserModel struct {
    ID            uint        `json:"id" gorm:"primary_key"`
    Username      string      `json:"username" gorm:"column:username"`
    Email         string      `json:"email" gorm:"column:email;unique_index"`
    Bio           string      `json:"bio" gorm:"column:bio"`
    Image         *string     `json:"image" gorm:"column:image"`
    PasswordHash  string      `json:"-" gorm:"column:password"`
    JWT           string      `json:"jwt" gorm:"column:-"`
}

func (u *UserModel) setPassword(password string){
    bytePassword := []byte(password)
    passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    u.PasswordHash = string(passwordHash)
}

func (u *UserModel) checkPassword(password string) error{
    bytePassword := []byte(password)
    byteHashedPassword := []byte(u.PasswordHash)
    token, err := common.GenToken(u.ID)
    if err!=nil {
        return err
    }
    u.JWT = token
    return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}