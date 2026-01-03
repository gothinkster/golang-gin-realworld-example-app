package articles

import (
	"strconv"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"gorm.io/gorm"
)

type ArticleModel struct {
	gorm.Model
	Slug        string `gorm:"uniqueIndex"`
	Title       string
	Description string `gorm:"size:2048"`
	Body        string `gorm:"size:2048"`
	Author      ArticleUserModel
	AuthorID    uint
	Tags        []TagModel     `gorm:"many2many:article_tags;"`
	Comments    []CommentModel `gorm:"ForeignKey:ArticleID"`
}

type ArticleUserModel struct {
	gorm.Model
	UserModel      users.UserModel
	UserModelID    uint
	ArticleModels  []ArticleModel  `gorm:"ForeignKey:AuthorID"`
	FavoriteModels []FavoriteModel `gorm:"ForeignKey:FavoriteByID"`
}

type FavoriteModel struct {
	gorm.Model
	Favorite     ArticleModel
	FavoriteID   uint
	FavoriteBy   ArticleUserModel
	FavoriteByID uint
}

type TagModel struct {
	gorm.Model
	Tag           string         `gorm:"uniqueIndex"`
	ArticleModels []ArticleModel `gorm:"many2many:article_tags;"`
}

type CommentModel struct {
	gorm.Model
	Article   ArticleModel
	ArticleID uint
	Author    ArticleUserModel
	AuthorID  uint
	Body      string `gorm:"size:2048"`
}

func GetArticleUserModel(userModel users.UserModel) ArticleUserModel {
	var articleUserModel ArticleUserModel
	if userModel.ID == 0 {
		return articleUserModel
	}
	db := common.GetDB()
	db.Where(&ArticleUserModel{
		UserModelID: userModel.ID,
	}).FirstOrCreate(&articleUserModel)
	articleUserModel.UserModel = userModel
	return articleUserModel
}

func (article ArticleModel) favoritesCount() uint {
	db := common.GetDB()
	var count int64
	db.Model(&FavoriteModel{}).Where(FavoriteModel{
		FavoriteID: article.ID,
	}).Count(&count)
	if count < 0 {
		return 0
	}
	return uint(count)
}

func (article ArticleModel) isFavoriteBy(user ArticleUserModel) bool {
	db := common.GetDB()
	var favorite FavoriteModel
	db.Where(FavoriteModel{
		FavoriteID:   article.ID,
		FavoriteByID: user.ID,
	}).First(&favorite)
	return favorite.ID != 0
}

func (article ArticleModel) favoriteBy(user ArticleUserModel) error {
	db := common.GetDB()
	var favorite FavoriteModel
	err := db.FirstOrCreate(&favorite, &FavoriteModel{
		FavoriteID:   article.ID,
		FavoriteByID: user.ID,
	}).Error
	return err
}

func (article ArticleModel) unFavoriteBy(user ArticleUserModel) error {
	db := common.GetDB()
	err := db.Where("favorite_id = ? AND favorite_by_id = ?", article.ID, user.ID).Delete(&FavoriteModel{}).Error
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
	err := db.Preload("Author.UserModel").Preload("Tags").Where(condition).First(&model).Error
	return model, err
}

func (self *ArticleModel) getComments() error {
	db := common.GetDB()
	err := db.Preload("Author.UserModel").Model(self).Association("Comments").Find(&self.Comments)
	return err
}

func getAllTags() ([]TagModel, error) {
	db := common.GetDB()
	var models []TagModel
	err := db.Find(&models).Error
	return models, err
}

func FindManyArticle(tag, author, limit, offset, favorited string) ([]ArticleModel, int, error) {
	db := common.GetDB()
	var models []ArticleModel
	var count int

	offset_int, errOffset := strconv.Atoi(offset)
	if errOffset != nil {
		offset_int = 0
	}

	limit_int, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		limit_int = 20
	}

	tx := db.Begin()
	if tag != "" {
		var tagModel TagModel
		tx.Where(TagModel{Tag: tag}).First(&tagModel)
		if tagModel.ID != 0 {
			// Get article IDs via association
			var tempModels []ArticleModel
			if err := tx.Model(&tagModel).Offset(offset_int).Limit(limit_int).Association("ArticleModels").Find(&tempModels); err != nil {
				tx.Rollback()
				return models, count, err
			}
			count = int(tx.Model(&tagModel).Association("ArticleModels").Count())
			// Fetch articles with preloaded associations in single query, ordered by updated_at desc
			if len(tempModels) > 0 {
				var ids []uint
				for _, m := range tempModels {
					ids = append(ids, m.ID)
				}
				tx.Preload("Author.UserModel").Preload("Tags").Where("id IN ?", ids).Order("updated_at desc").Find(&models)
			}
		}
	} else if author != "" {
		var userModel users.UserModel
		tx.Where(users.UserModel{Username: author}).First(&userModel)
		articleUserModel := GetArticleUserModel(userModel)

		if articleUserModel.ID != 0 {
			count = int(tx.Model(&articleUserModel).Association("ArticleModels").Count())
			// Get article IDs via association
			var tempModels []ArticleModel
			if err := tx.Model(&articleUserModel).Offset(offset_int).Limit(limit_int).Association("ArticleModels").Find(&tempModels); err != nil {
				tx.Rollback()
				return models, count, err
			}
			// Fetch articles with preloaded associations in single query, ordered by updated_at desc
			if len(tempModels) > 0 {
				var ids []uint
				for _, m := range tempModels {
					ids = append(ids, m.ID)
				}
				tx.Preload("Author.UserModel").Preload("Tags").Where("id IN ?", ids).Order("updated_at desc").Find(&models)
			}
		}
	} else if favorited != "" {
		var userModel users.UserModel
		tx.Where(users.UserModel{Username: favorited}).First(&userModel)
		articleUserModel := GetArticleUserModel(userModel)
		if articleUserModel.ID != 0 {
			var favoriteModels []FavoriteModel
			tx.Where(FavoriteModel{
				FavoriteByID: articleUserModel.ID,
			}).Offset(offset_int).Limit(limit_int).Find(&favoriteModels)

			count = int(tx.Model(&articleUserModel).Association("FavoriteModels").Count())
			for _, favorite := range favoriteModels {
				var model ArticleModel
				tx.Preload("Author.UserModel").Preload("Tags").First(&model, favorite.FavoriteID)
				models = append(models, model)
			}
		}
	} else {
		var count64 int64
		tx.Model(&ArticleModel{}).Count(&count64)
		count = int(count64)
		tx.Offset(offset_int).Limit(limit_int).Preload("Author.UserModel").Preload("Tags").Find(&models)
	}

	err := tx.Commit().Error
	return models, count, err
}

func (self *ArticleUserModel) GetArticleFeed(limit, offset string) ([]ArticleModel, int, error) {
	db := common.GetDB()
	var models []ArticleModel
	var count int

	offset_int, errOffset := strconv.Atoi(offset)
	if errOffset != nil {
		offset_int = 0
	}
	limit_int, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		limit_int = 20
	}

	tx := db.Begin()
	followings := self.UserModel.GetFollowings()
	var articleUserModels []uint
	for _, following := range followings {
		articleUserModel := GetArticleUserModel(following)
		articleUserModels = append(articleUserModels, articleUserModel.ID)
	}

	if len(articleUserModels) > 0 {
		var count64 int64
		tx.Model(&ArticleModel{}).Where("author_id in (?)", articleUserModels).Count(&count64)
		count = int(count64)
		tx.Preload("Author.UserModel").Preload("Tags").Where("author_id in (?)", articleUserModels).Order("updated_at desc").Offset(offset_int).Limit(limit_int).Find(&models)
	}

	err := tx.Commit().Error
	if models == nil {
		models = []ArticleModel{}
	}
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
	err := db.Model(model).Updates(data).Error
	return err
}

func DeleteArticleModel(condition interface{}) error {
	db := common.GetDB()
	err := db.Where(condition).Delete(&ArticleModel{}).Error
	return err
}

func DeleteCommentModel(condition interface{}) error {
	db := common.GetDB()
	err := db.Where(condition).Delete(&CommentModel{}).Error
	return err
}
