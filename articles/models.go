package articles

import (
	"fmt"
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

// BatchGetFavoriteCounts returns a map of article ID to favorite count
func BatchGetFavoriteCounts(articleIDs []uint) map[uint]uint {
	if len(articleIDs) == 0 {
		return make(map[uint]uint)
	}
	db := common.GetDB()

	type result struct {
		FavoriteID uint
		Count      uint
	}
	var results []result
	db.Model(&FavoriteModel{}).
		Select("favorite_id, COUNT(*) as count").
		Where("favorite_id IN ?", articleIDs).
		Group("favorite_id").
		Find(&results)

	countMap := make(map[uint]uint)
	for _, r := range results {
		countMap[r.FavoriteID] = r.Count
	}
	return countMap
}

// BatchGetFavoriteStatus returns a map of article ID to whether the user favorited it
func BatchGetFavoriteStatus(articleIDs []uint, userID uint) map[uint]bool {
	if len(articleIDs) == 0 || userID == 0 {
		return make(map[uint]bool)
	}
	db := common.GetDB()

	var favorites []FavoriteModel
	db.Where("favorite_id IN ? AND favorite_by_id = ?", articleIDs, userID).Find(&favorites)

	statusMap := make(map[uint]bool)
	for _, f := range favorites {
		statusMap[f.FavoriteID] = true
	}
	return statusMap
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

func FindOneComment(condition *CommentModel) (CommentModel, error) {
	db := common.GetDB()
	var model CommentModel
	err := db.Preload("Author.UserModel").Preload("Article").Where(condition).First(&model).Error
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
	models := make([]ArticleModel, 0)

	offset_int, errOffset := strconv.Atoi(offset)
	if errOffset != nil {
		offset_int = 0
	}

	limit_int, errLimit := strconv.Atoi(limit)
	if errLimit != nil {
		limit_int = 20
	}

	// buildQuery returns a fresh filtered query so it can be used once for the
	// total count and once for the paginated fetch. The bool is false when the
	// filter target (tag or user) does not exist, meaning an empty result.
	buildQuery := func() (*gorm.DB, bool) {
		query := db.Model(&ArticleModel{})
		if tag != "" {
			var tagModel TagModel
			if err := db.Where(TagModel{Tag: tag}).First(&tagModel).Error; err != nil {
				return nil, false
			}
			query = query.
				Joins("JOIN article_tags ON article_tags.article_model_id = article_models.id").
				Where("article_tags.tag_model_id = ?", tagModel.ID)
		} else if author != "" {
			var userModel users.UserModel
			if err := db.Where(users.UserModel{Username: author}).First(&userModel).Error; err != nil {
				return nil, false
			}
			var articleUserModel ArticleUserModel
			if err := db.Where(ArticleUserModel{UserModelID: userModel.ID}).First(&articleUserModel).Error; err != nil {
				return nil, false
			}
			query = query.Where("author_id = ?", articleUserModel.ID)
		} else if favorited != "" {
			var userModel users.UserModel
			if err := db.Where(users.UserModel{Username: favorited}).First(&userModel).Error; err != nil {
				return nil, false
			}
			var articleUserModel ArticleUserModel
			if err := db.Where(ArticleUserModel{UserModelID: userModel.ID}).First(&articleUserModel).Error; err != nil {
				return nil, false
			}
			query = query.
				Joins("JOIN favorite_models ON favorite_models.favorite_id = article_models.id").
				Where("favorite_models.favorite_by_id = ? AND favorite_models.deleted_at IS NULL", articleUserModel.ID)
		}
		return query, true
	}

	countQuery, ok := buildQuery()
	if !ok {
		return models, 0, nil
	}
	var count64 int64
	if err := countQuery.Count(&count64).Error; err != nil {
		return models, 0, err
	}

	fetchQuery, ok := buildQuery()
	if !ok {
		return models, 0, nil
	}
	err := fetchQuery.
		Preload("Author.UserModel").Preload("Tags").
		Order("article_models.created_at DESC").
		Offset(offset_int).Limit(limit_int).
		Find(&models).Error
	return models, int(count64), err
}

func (self *ArticleUserModel) GetArticleFeed(limit, offset string) ([]ArticleModel, int, error) {
	db := common.GetDB()
	models := make([]ArticleModel, 0)
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

	// Batch get ArticleUserModel IDs to avoid N+1 query
	if len(followings) > 0 {
		var followingUserIDs []uint
		for _, following := range followings {
			followingUserIDs = append(followingUserIDs, following.ID)
		}

		var articleUserModels []ArticleUserModel
		tx.Where("user_model_id IN ?", followingUserIDs).Find(&articleUserModels)

		var authorIDs []uint
		for _, aum := range articleUserModels {
			authorIDs = append(authorIDs, aum.ID)
		}

		if len(authorIDs) > 0 {
			var count64 int64
			tx.Model(&ArticleModel{}).Where("author_id IN ?", authorIDs).Count(&count64)
			count = int(count64)
			tx.Preload("Author.UserModel").Preload("Tags").Where("author_id IN ?", authorIDs).Order("created_at desc").Offset(offset_int).Limit(limit_int).Find(&models)
		}
	}

	err := tx.Commit().Error
	return models, count, err
}

// makeUniqueSlug returns baseSlug, or baseSlug-N for the smallest N that does
// not collide with any existing article (including soft-deleted rows, which
// still occupy the unique index).
func makeUniqueSlug(baseSlug string) string {
	db := common.GetDB()
	candidate := baseSlug
	for suffix := 2; ; suffix++ {
		var count int64
		db.Unscoped().Model(&ArticleModel{}).Where("slug = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}
}

// ReplaceTags persists a new tag set for an existing article, creating any
// missing tags and detaching the ones no longer referenced.
func (model *ArticleModel) ReplaceTags(tags []string) error {
	tagList, err := buildTagModels(tags)
	if err != nil {
		return err
	}
	db := common.GetDB()
	if err := db.Model(model).Association("Tags").Replace(&tagList); err != nil {
		return err
	}
	model.Tags = tagList
	return nil
}

func (model *ArticleModel) setTags(tags []string) error {
	tagList, err := buildTagModels(tags)
	if err != nil {
		return err
	}
	model.Tags = tagList
	return nil
}

func buildTagModels(tags []string) ([]TagModel, error) {
	if len(tags) == 0 {
		return []TagModel{}, nil
	}

	db := common.GetDB()

	// Batch fetch existing tags
	var existingTags []TagModel
	db.Where("tag IN ?", tags).Find(&existingTags)

	// Create a map for quick lookup
	existingTagMap := make(map[string]TagModel)
	for _, t := range existingTags {
		existingTagMap[t.Tag] = t
	}

	// Create missing tags and build final list
	var tagList []TagModel
	for _, tag := range tags {
		if existing, ok := existingTagMap[tag]; ok {
			tagList = append(tagList, existing)
		} else {
			// Create new tag with race condition handling
			newTag := TagModel{Tag: tag}
			if err := db.Create(&newTag).Error; err != nil {
				// If creation failed (e.g., concurrent insert), try to fetch existing
				var existing TagModel
				if err2 := db.Where("tag = ?", tag).First(&existing).Error; err2 == nil {
					tagList = append(tagList, existing)
					continue
				}
				return nil, err
			}
			tagList = append(tagList, newTag)
		}
	}
	return tagList, nil
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
