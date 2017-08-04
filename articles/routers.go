package articles

import (
    "golang-gin-starter-kit/common"
    "gopkg.in/gin-gonic/gin.v1"
    "net/http"
    "errors"
    "golang-gin-starter-kit/users"
)

func ArticlesRegister(router *gin.RouterGroup) {
    router.GET("/", ArticleList)
    router.POST("/", ArticleCreate)
    router.GET("/:slug", ArticleRetrieve)
    router.PUT("/:slug", ArticleUpdate)
    router.DELETE("/:slug", ArticleDelete)
    router.POST("/:slug/favorite", ArticleFavorite)
    router.DELETE("/:slug/favorite", ArticleUnfavorite)
}

func ArticleCreate(c *gin.Context) {
    articleModelValidator := NewArticleModelValidator()
    if err := articleModelValidator.Bind(c); err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
        return
    }

    if err := SaveOne(&articleModelValidator.articleModel); err != nil {
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
    articleModels, modelCount, err := FindManyArticle(tag,author,limit,offset,favorited)
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid param")))
        return
    }
    serializer := ArticlesSerializer{c, articleModels}
    c.JSON(http.StatusOK, gin.H{"articles": serializer.Response(), "articlesCount":modelCount})
}

func ArticleRetrieve(c *gin.Context) {
    slug := c.Param("slug")
    articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid slug")))
        return
    }
    serializer := ArticleSerializer{c, articleModel}
    c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleUpdate(c *gin.Context) {
    slug := c.Param("slug")
    articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid slug")))
        return
    }
    articleModelValidator := NewArticleModelValidatorFillWith(articleModel)
    if err := articleModelValidator.Bind(c); err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))
        return
    }

    if err := articleModel.Update(articleModelValidator.articleModel); err != nil {
        c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
        return
    }
    serializer := ArticleSerializer{c, articleModel}
    c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleDelete(c *gin.Context) {
    slug := c.Param("slug")
    err := DeleteArticleModel(&ArticleModel{Slug: slug})
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid slug")))
        return
    }
    c.JSON(http.StatusOK, gin.H{"article": "Delete success"})
}

func ArticleFavorite(c *gin.Context) {
    slug := c.Param("slug")
    articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid slug")))
        return
    }
    myUserModel := c.MustGet("my_user_model").(users.UserModel)
    err = articleModel.favoriteBy(myUserModel)
    serializer := ArticleSerializer{c, articleModel}
    c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}

func ArticleUnfavorite(c *gin.Context) {
    slug := c.Param("slug")
    articleModel, err := FindOneArticle(&ArticleModel{Slug: slug})
    if err != nil {
        c.JSON(http.StatusNotFound, common.NewError("articles", errors.New("Invalid slug")))
        return
    }
    myUserModel := c.MustGet("my_user_model").(users.UserModel)
    err = articleModel.unFavoriteBy(myUserModel)
    serializer := ArticleSerializer{c, articleModel}
    c.JSON(http.StatusOK, gin.H{"article": serializer.Response()})
}
