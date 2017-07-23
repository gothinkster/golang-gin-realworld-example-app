package users

import (
    "golang.org/x/crypto/bcrypt"
    "golang-gin-starter-kit/common"
    "errors"
)

type UserModel struct {
    ID            uint64      `json:"id" gorm:"primary_key"`
    Username      string      `json:"username" gorm:"column:username"`
    Email         string      `json:"email" gorm:"column:email;unique_index"`
    Bio           string      `json:"bio" gorm:"column:bio"`
    Image         *string     `json:"image" gorm:"column:image"`
    PasswordHash  string      `json:"-" gorm:"column:password;not null"`
    JWT           string      `json:"jwt" gorm:"column:-"`
}

func (u *UserModel) setJWT()error{
    token, err := common.GenToken(u.ID)
    u.JWT = token
    return err
}

func (u *UserModel) setPassword(password string) error{
    if len(password)==0{
        return errors.New("password should not be empty!")
    }
    bytePassword := []byte(password)
    passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    if err!=nil {
        return err
    }
    u.PasswordHash = string(passwordHash)
    return nil
}

func (u *UserModel) checkPassword(password string) error{
    bytePassword := []byte(password)
    byteHashedPassword := []byte(u.PasswordHash)
    return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}