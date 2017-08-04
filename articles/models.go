package articles

import (
    "github.com/jinzhu/gorm"
    "golang-gin-starter-kit/common"
    "golang-gin-starter-kit/users"
    "fmt"
    "strconv"
)

type ArticleModel struct {
    gorm.Model
    Slug            string              `gorm:"unique_index"`
    Title           string
    Description     string              `gorm:"size:2048"`
    Body            string              `gorm:"size:2048"`
    Author          users.UserModel
    AuthorID        uint
    Tags            []TagModel          `gorm:"many2many:article_tags;"`
}

type FavoriteModel struct {
    gorm.Model
    Favorite     ArticleModel
    FavoriteID   uint
    FavoriteBy   users.UserModel
    FavoriteByID uint
}

type TagModel struct {
    gorm.Model
    Tag             string          `gorm:"unique_index"`
    ArticleModels   []ArticleModel  `gorm:"many2many:article_tags;"`
}

func (article ArticleModel) favoritesCount() uint {
    db := common.GetDB()
    var count uint
    db.Model(&FavoriteModel{}).Where(FavoriteModel{
        FavoriteID:   article.ID,
    }).Count(&count)
    return count
}

func (article ArticleModel) isFavoriteBy(user users.UserModel) bool {
    db := common.GetDB()
    var favorite FavoriteModel
    db.Where(FavoriteModel{
        FavoriteID:   article.ID,
        FavoriteByID: user.ID,
    }).First(&favorite)
    return favorite.ID != 0
}

func (article ArticleModel) favoriteBy(user users.UserModel) error {
    db := common.GetDB()
    var favorite FavoriteModel
    err := db.FirstOrCreate(&favorite, &FavoriteModel{
        FavoriteID:   article.ID,
        FavoriteByID: user.ID,
    }).Error
    return err
}

func (article ArticleModel) unFavoriteBy(user users.UserModel) error {
    db := common.GetDB()
    err := db.Where(FavoriteModel{
        FavoriteID:   article.ID,
        FavoriteByID: user.ID,
    }).Delete(FavoriteModel{}).Error
    return err
}

func SaveOne(data interface{}) error {
    db := common.GetDB()
    err := db.Save(data).Error
    return err
}

func FindOneArticle(condition interface{}) (ArticleModel, error) {
    db := common.GetDB()
    var model ArticleModel
    tx := db.Begin()
    tx.Where(condition).First(&model)
    tx.Model(&model).Related(&model.Author,"Author")
    tx.Model(&model).Related(&model.Tags,"Tags")
    err := tx.Commit().Error
    return model, err
}

func FindManyArticle(tag, author, limit, offset, favorited string) ([]ArticleModel,int, error) {
    db := common.GetDB()
    var models []ArticleModel
    var count int
    tx := db.Begin()
    if tag!="" {
        var tagModel TagModel
        tx.Where(TagModel{Tag:tag}).First(&tagModel)
        tx.Model(&tagModel).Related(&models,"ArticleModels")
    }else if author!="" {
        var models_tmp []ArticleModel
        var userModel users.UserModel
        tx.Model(&users.UserModel{Username:author}).First(&userModel)
        tx.Where(&ArticleModel{Author:userModel}).Find(models_tmp)
        models = models_tmp
    }
    fmt.Println("1gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg")
    fmt.Println(models)

    if favorited!="" {
        var favoriteModels []FavoriteModel
        var userModel users.UserModel
        tx.Model(&users.UserModel{Username:author}).First(&userModel)
        db.Where(FavoriteModel{
            FavoriteByID:  userModel.ID,
        }).Find(&favoriteModels)

        fmt.Println("0gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg")
        fmt.Println(favoriteModels)
    }
    fmt.Println("2gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg")
    fmt.Println(models)
    tx.Model(&models).Count(&count)
    offset_int, err := strconv.Atoi(offset)
    if err!=nil{
        offset_int = 0
    }

    limit_int, err := strconv.Atoi(limit)
    if err!=nil{
        limit_int = 20
    }

    tx.Model(&models).Offset(offset_int).Limit(limit_int).Find(&models)
    for i, _ := range models{
        tx.Model(&models[i]).Related(&models[i].Author,"Author")
        tx.Model(&models[i]).Related(&models[i].Tags,"Tags")
    }
    err = tx.Commit().Error
    return models, count, err
}

func (model *ArticleModel) setTags(tags []string) error {
    db := common.GetDB()
    var tagList []TagModel
    for _, tag := range tags {
        var tagModel TagModel
        err := db.FirstOrCreate(&tagModel, TagModel{Tag: tag}).Error
        if err != nil {
            return err
        }
        tagList = append(tagList, tagModel)
    }
    model.Tags = tagList
    return nil
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
