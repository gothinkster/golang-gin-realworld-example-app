package articles

import (
    "golang-gin-starter-kit/common"
    "gopkg.in/gin-gonic/gin.v1"
    "net/http"
)

func ArticlesRegister(router *gin.RouterGroup) {
    router.POST("/", ArticleCreate)
    router.PUT("/:slug", ArticleUpdate)
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
    articleSerializer := ArticleSerializer{c,articleModelValidator.articleModel}
    c.JSON(http.StatusCreated, gin.H{"article": articleSerializer.Response()})
}


func ArticleUpdate(c *gin.Context) {
}
