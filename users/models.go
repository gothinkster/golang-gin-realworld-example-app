package users

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"golang.org/x/crypto/bcrypt"
)

// Models should only be concerned with database schema, more strict checking should be put in validator.
//
// More detail you can find here: http://jinzhu.me/gorm/models.html#model-definition
//
// HINT: If you want to split null and "", you should use *string instead of string.
type UserModel struct {
	ID           uint    `gorm:"primary_key"`
	Username     string  `gorm:"column:username"`
	Email        string  `gorm:"column:email;unique_index"`
	Bio          string  `gorm:"column:bio;size:1024"`
	Image        *string `gorm:"column:image"`
	PasswordHash string  `gorm:"column:password;not null"`
}

// A hack way to save ManyToMany relationship,
// gorm will build the alias as FollowingBy <-> FollowingByID <-> "following_by_id".
//
// DB schema looks like: id, created_at, updated_at, deleted_at, following_id, followed_by_id.
//
// Retrieve them by:
// 	db.Where(FollowModel{ FollowingID:  v.ID, FollowedByID: u.ID, }).First(&follow)
// 	db.Where(FollowModel{ FollowedByID: u.ID, }).Find(&follows)
//
// More details about gorm.Model: http://jinzhu.me/gorm/models.html#conventions
type FollowModel struct {
	gorm.Model
	Following    UserModel
	FollowingID  uint
	FollowedBy   UserModel
	FollowedByID uint
}

// Migrate the schema of database if needed
func AutoMigrate() {
	db := common.GetDB()

	db.AutoMigrate(&UserModel{})
	db.AutoMigrate(&FollowModel{})
}

// What's bcrypt? https://en.wikipedia.org/wiki/Bcrypt
// Golang bcrypt doc: https://godoc.org/golang.org/x/crypto/bcrypt
// You can change the value in bcrypt.DefaultCost to adjust the security index.
// 	err := userModel.setPassword("password0")
func (u *UserModel) setPassword(password string) error {
	if len(password) == 0 {
		return errors.New("password should not be empty!")
	}
	bytePassword := []byte(password)
	// Make sure the second param `bcrypt generator cost` between [4, 32)
	passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	u.PasswordHash = string(passwordHash)
	return nil
}

// Database will only save the hashed string, you should check it by util function.
// 	if err := serModel.checkPassword("password0"); err != nil { password error }
func (u *UserModel) checkPassword(password string) error {
	bytePassword := []byte(password)
	byteHashedPassword := []byte(u.PasswordHash)
	return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}

// You could input the conditions and it will return an UserModel in database with error info.
// 	userModel, err := FindOneUser(&UserModel{Username: "username0"})
func FindOneUser(condition interface{}) (UserModel, error) {
	db := common.GetDB()
	var model UserModel
	err := db.Where(condition).First(&model).Error
	return model, err
}

// You could input an UserModel which will be saved in database returning with error info
// 	if err := SaveOne(&userModel); err != nil { ... }
func SaveOne(data interface{}) error {
	db := common.GetDB()
	err := db.Save(data).Error
	return err
}

// You could update properties of an UserModel to database returning with error info.
//  err := db.Model(userModel).Update(UserModel{Username: "wangzitian0"}).Error
func (model *UserModel) Update(data interface{}) error {
	db := common.GetDB()
	err := db.Model(model).Update(data).Error
	return err
}

// You could add a following relationship as userModel1 following userModel2
// 	err = userModel1.following(userModel2)
func (u UserModel) following(v UserModel) error {
	db := common.GetDB()
	var follow FollowModel
	err := db.FirstOrCreate(&follow, &FollowModel{
		FollowingID:  v.ID,
		FollowedByID: u.ID,
	}).Error
	return err
}

// You could check whether  userModel1 following userModel2
// 	followingBool = myUserModel.isFollowing(self.UserModel)
func (u UserModel) isFollowing(v UserModel) bool {
	db := common.GetDB()
	var follow FollowModel
	db.Where(FollowModel{
		FollowingID:  v.ID,
		FollowedByID: u.ID,
	}).First(&follow)
	return follow.ID != 0
}

// You could delete a following relationship as userModel1 following userModel2
// 	err = userModel1.unFollowing(userModel2)
func (u UserModel) unFollowing(v UserModel) error {
	db := common.GetDB()
	err := db.Where(FollowModel{
		FollowingID:  v.ID,
		FollowedByID: u.ID,
	}).Delete(FollowModel{}).Error
	return err
}

// You could get a following list of userModel
// 	followings := userModel.GetFollowings()
func (u UserModel) GetFollowings() []UserModel {
	db := common.GetDB()
	tx := db.Begin()
	var follows []FollowModel
	var followings []UserModel
	tx.Where(FollowModel{
		FollowedByID: u.ID,
	}).Find(&follows)
	for _, follow := range follows {
		var userModel UserModel
		tx.Model(&follow).Related(&userModel, "Following")
		followings = append(followings, userModel)
	}
	tx.Commit()
	return followings
}
