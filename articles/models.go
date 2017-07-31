package articles

import (
    "github.com/jinzhu/gorm"
    "golang-gin-starter-kit/common"
    "golang-gin-starter-kit/users"
)

type ArticleModel struct {
    gorm.Model
    Title           string      `gorm:"unique_index"`
    Description     string      `gorm:"size:2048"`
    Body            string      `gorm:"size:2048"`
    Author          users.UserModel
}


func FindOneArticle(condition interface{}) (ArticleModel, error) {
    db := common.GetDB()
    var model ArticleModel
    err := db.Where(condition).First(&model).Error
    return model, err
}

func SaveOne(data interface{}) error {
    db := common.GetDB()
    err := db.Save(data).Error
    return err
}

func (model *ArticleModel) Update(data interface{}) error {
    db := common.GetDB()
    err := db.Model(model).Update(data).Error
    return err
}


func DeleteArticleModel(condition interface{}) error {
    db := common.GetDB()
    err := db.Where(condition).Delete(ArticleModel{}).Error
    return err
}
