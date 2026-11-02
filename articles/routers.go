package articles

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"gorm.io/gorm"
)

func ArticlesRegister(router *gin.RouterGroup) {
	router.GET("/feed", ArticleFeed)
	router.POST("", ArticleCreate)
	router.POST("/", ArticleCreate)
	router.PUT("/:slug", ArticleUpdate)
	router.PUT("/:slug/", ArticleUpdate)
	router.DELETE("/:slug", ArticleDelete)
	router.POST("/:slug/favorite", ArticleFavorite)
	router.DELETE("/:slug/favorite", ArticleUnfavorite)
	router.POST("/:slug/comments", ArticleCommentCreate)
	router.DELETE("/:slug/comments/:id", ArticleCommentDelete)
}

func ArticlesAnonymousRegister(router *gin.RouterGroup) {
	router.GET("", ArticleList)
	router.GET("/", ArticleList)
	router.GET("/:slug", ArticleRetrieve)
	router.GET("/:slug/comments", ArticleCommentList)
}

func TagsAnonymousRegister(router *gin.RouterGroup) {
	router.GET("", TagList)
	router.GET("/", TagList)
}

func ArticleCreate(c *gin.Context) {
	articleModelValidator := NewArticleModelValidator()
	if err := articleModelValidator.Bind(c); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	// Duplicate titles are allowed; each article still needs a unique slug.
	// makeUniqueSlug races with concurrent creates of the same title, so on a
	// unique-index violation regenerate and retry a couple of times.
	baseSlug := articleModelValidator.articleModel.Slug
	for attempt := 0; ; attempt++ {
		articleModelValidator.articleModel.Slug = makeUniqueSlug(baseSlug)
		err := SaveOne(&articleModelValidator.articleModel)
		if err == nil {
			break
		}
		if attempt < 2 && errors.Is(err, gorm.ErrDuplicatedKey) {
			continue
		}
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ArticleSerializer{c, articleModelValidator.articleModel}
	c.JSON(http.StatusCreated, gin.H{"article": serializer.Response()})
}

func ArticleList(c *gin.Context) {
	//condition := ArticleModel{}
	tag := c.Query("tag")
	author := c.Query("author")
	favorited := c.Query("favorited")
	limit := c.Query("limit")
	offset := c.Query("offset")
	articleModels, modelCount, err := FindManyArticle(tag, author, limit, offset, favorited)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid param")))
		return
	}
	serializer := ArticlesSerializer{c, articleModels}
	c.JSON(http.StatusOK, gin.H{"articles": serializer.Response(), "articlesCount": modelCount})
}

func ArticleFeed(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	if myUserModel.ID == 0 {
		c.AbortWithError(http.StatusUnauthorized, errors.New("{error : \"Require auth!\"}"))
		return
	}
	articleUserModel := GetArticleUserModel(myUserModel)
	articleModels, modelCount, err := articleUserModel.GetArticleFeed(limit, offset)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid param")))
		return
	}
	serializer := ArticlesSerializer{c, articleModels}
	c.JSON(http.StatusOK, gin.H{"articles": serializer.Response(), "articlesCount": modelCount})
}

func ArticleRetrieve(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	serializer := ArticleSerializer{c, articleModel}
	c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

// ArticleUpdate handles PUT /api/articles/:slug. The schema's Nullable fields
// preserve omitted keys (including tagList) and reject explicit null with 422;
// validation runs declaratively via the binding tags during Bind.
func ArticleUpdate(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	// Check if current user is the author
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	articleUserModel := GetArticleUserModel(myUserModel)
	if articleModel.AuthorID != articleUserModel.ID {
		c.JSON(http.StatusForbidden, common.NewErrorMessage("article", "forbidden"))
		return
	}

	articleUpdateValidator := NewArticleUpdateValidator()
	if err := articleUpdateValidator.Bind(c); err != nil {
		errs := common.NewValidatorError(err)
		errs.MarkInvalidFields(articleUpdateValidator.invalidFields())
		c.JSON(http.StatusUnprocessableEntity, errs)
		return
	}
	// Past validation, every Set field holds a valid value.
	var newTags *[]string
	if f := articleUpdateValidator.Article.TagList; f.Set {
		tags := f.Value
		newTags = &tags
	}
	updates := map[string]interface{}{}
	if f := articleUpdateValidator.Article.Title; f.Set {
		updates["title"] = f.Value
	}
	if f := articleUpdateValidator.Article.Description; f.Set {
		updates["description"] = f.Value
	}
	if f := articleUpdateValidator.Article.Body; f.Set {
		updates["body"] = f.Value
	}

	// Changing the title regenerates the slug, as in the original RealWorld backends.
	finalSlug := slug
	newSlugBase := ""
	if title, ok := updates["title"]; ok {
		if base := makeSlug(title.(string)); base != articleModel.Slug {
			newSlugBase = base
			finalSlug = makeUniqueSlug(base)
			updates["slug"] = finalSlug
		}
	}
	if len(updates) > 0 {
		// Same slug race as on create: regenerate and retry on collision.
		for attempt := 0; ; attempt++ {
			err := articleModel.Update(updates)
			if err == nil {
				break
			}
			if attempt < 2 && newSlugBase != "" && errors.Is(err, gorm.ErrDuplicatedKey) {
				finalSlug = makeUniqueSlug(newSlugBase)
				updates["slug"] = finalSlug
				continue
			}
			c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
			return
		}
	}
	if newTags != nil {
		if err := articleModel.ReplaceTags(*newTags); err != nil {
			c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
			return
		}
	}

	articleModel, err = FindOneArticle(&ArticleModel{Slug: finalSlug})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ArticleSerializer{c, articleModel}
	c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleDelete(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	articleUserModel := GetArticleUserModel(myUserModel)
	if articleModel.AuthorID != articleUserModel.ID {
		c.JSON(http.StatusForbidden, common.NewErrorMessage("article", "forbidden"))
		return
	}
	if err := DeleteArticleModel(&ArticleModel{Slug: slug}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	c.Status(http.StatusNoContent)
}

func ArticleFavorite(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	if err = articleModel.favoriteBy(GetArticleUserModel(myUserModel)); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ArticleSerializer{c, articleModel}
	c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleUnfavorite(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	if err = articleModel.unFavoriteBy(GetArticleUserModel(myUserModel)); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := ArticleSerializer{c, articleModel}
	c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleCommentCreate(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	commentModelValidator := NewCommentModelValidator()
	if err := commentModelValidator.Bind(c); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
		return
	}
	commentModelValidator.commentModel.Article = articleModel

	if err := SaveOne(&commentModelValidator.commentModel); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	serializer := CommentSerializer{c, commentModelValidator.commentModel}
	c.JSON(http.StatusCreated, gin.H{"comment": serializer.Response()})
}

func ArticleCommentDelete(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("comment", "not found"))
		return
	}
	id := uint(id64)
	commentModel, err := FindOneComment(&CommentModel{Model: gorm.Model{ID: id}, ArticleID: articleModel.ID})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("comment", "not found"))
		return
	}
	myUserModel := c.MustGet("my_user_model").(users.UserModel)
	articleUserModel := GetArticleUserModel(myUserModel)
	if commentModel.AuthorID != articleUserModel.ID {
		c.JSON(http.StatusForbidden, common.NewErrorMessage("comment", "forbidden"))
		return
	}
	if err := DeleteCommentModel([]uint{id}); err != nil {
		c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
		return
	}
	c.Status(http.StatusNoContent)
}

func ArticleCommentList(c *gin.Context) {
	slug := c.Param("slug")
	articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewErrorMessage("article", "not found"))
		return
	}
	err = articleModel.getComments()
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("comments", errors.New("Database error")))
		return
	}
	serializer := CommentsSerializer{c, articleModel.Comments}
	c.JSON(http.StatusOK, gin.H{"comments": serializer.Response()})
}
func TagList(c *gin.Context) {
	tagModels, err := getAllTags()
	if err != nil {
		c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid param")))
		return
	}
	serializer := TagsSerializer{c, tagModels}
	c.JSON(http.StatusOK, gin.H{"tags": serializer.Response()})
}
