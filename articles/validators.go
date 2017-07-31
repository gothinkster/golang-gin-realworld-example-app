package articles

import (
    "golang-gin-starter-kit/common"
    "gopkg.in/gin-gonic/gin.v1"
    "golang-gin-starter-kit/users"
)

type ArticleModelValidator struct {
    Article struct {
        Title       string      `form:"title" json:"title" binding:"exists,min=4"`
        Description string      `form:"description" json:"description" binding:"max=2048"`
        Body        string      `form:"body" json:"body" binding:"max=2048"`
    } `json:"article"`
    articleModel ArticleModel   `json:"-"`
}

func NewArticleModelValidator() ArticleModelValidator {
    return ArticleModelValidator{};
}

func (self *ArticleModelValidator) Bind(c *gin.Context) error {
    myUserModel := c.MustGet("my_user_model").(users.UserModel)

    err := common.Bind(c, self)
    if err != nil {
        return err
    }
    self.articleModel.Title = self.Article.Title
    self.articleModel.Description = self.Article.Description
    self.articleModel.Body = self.Article.Body
    self.articleModel.Author = myUserModel


    return nil
}
