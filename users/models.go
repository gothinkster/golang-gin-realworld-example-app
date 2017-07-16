package users

import "github.com/jinzhu/gorm"


type UserModel struct {
    gorm.Model
    Username      string      `gorm:"column:username"`
    Email         string      `gorm:"column:email"`
    Bio           string      `gorm:"column:bio"`
    Image         string      `gorm:"column:image"`
    Salt          string      `gorm:"column:salt"`
    PasswordHash  string      `gorm:"column:password"`
}

func (u *UserModel) setPassword(password string){
    passwordHash := password + "salt"
    u.PasswordHash = passwordHash
}

func (u *UserModel) checkPassword(password string) bool{
    passwordHash := password + "salt"
    return passwordHash == u.PasswordHash
}