package articles

import (
	"github.com/gosimple/slug"
	"golang-gin-starter-kit/common"
	"golang-gin-starter-kit/users"
	"gopkg.in/gin-gonic/gin.v1"
)

type ArticleModelValidator struct {
	Article struct {
		Title       string   `form:"title" json:"title" binding:"exists,min=4"`
		Description string   `form:"description" json:"description" binding:"max=2048"`
		Body        string   `form:"body" json:"body" binding:"max=2048"`
		Tags        []string `form:"tagList" json:"tagList"`
	} `json:"article"`
	articleModel ArticleModel `json:"-"`
}

func NewArticleModelValidator() ArticleModelValidator {
	return ArticleModelValidator{}
}

func NewArticleModelValidatorFillWith(articleModel ArticleModel) ArticleModelValidator {
	articleModelValidator := NewArticleModelValidator()
	articleModelValidator.Article.Title = articleModel.Title
	articleModelValidator.Article.Description = articleModel.Description
	articleModelValidator.Article.Body = articleModel.Body
	for _, tagModel := range articleModel.Tags {
		articleModelValidator.Article.Tags = append(articleModelValidator.Article.Tags, tagModel.Tag)
	}
	return articleModelValidator
}

func (self *ArticleModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, self)
	if err != nil {
		return err
	}
	self.articleModel.Slug = slug.Make(self.Article.Title)
	self.articleModel.Title = self.Article.Title
	self.articleModel.Description = self.Article.Description
	self.articleModel.Body = self.Article.Body
	self.articleModel.Author = GetArticleUserModel(myUserModel)
	self.articleModel.setTags(self.Article.Tags)
	return nil
}

type CommentModelValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"max=2048"`
	} `json:"comment"`
	commentModel CommentModel `json:"-"`
}

func NewCommentModelValidator() CommentModelValidator {
	return CommentModelValidator{}
}

func (self *CommentModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, self)
	if err != nil {
		return err
	}
	self.commentModel.Body = self.Comment.Body
	self.commentModel.Author = GetArticleUserModel(myUserModel)
	return nil
}
