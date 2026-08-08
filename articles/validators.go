package articles

import (
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
)

type ArticleModelValidator struct {
	Article struct {
		Title       string   `form:"title" json:"title" binding:"required,min=4"`
		Description string   `form:"description" json:"description" binding:"required,max=2048"`
		Body        string   `form:"body" json:"body" binding:"required,max=2048"`
		Tags        []string `form:"tagList" json:"tagList"`
	} `json:"article"`
	articleModel ArticleModel `json:"-"`
}

// makeSlug wraps slug.Make so handlers can use it without importing the slug
// package under a name that would shadow local variables.
func makeSlug(title string) string {
	return slug.Make(title)
}

func NewArticleModelValidator() ArticleModelValidator {
	return ArticleModelValidator{}
}

func (s *ArticleModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, s)
	if err != nil {
		return err
	}
	s.articleModel.Slug = slug.Make(s.Article.Title)
	s.articleModel.Title = s.Article.Title
	s.articleModel.Description = s.Article.Description
	s.articleModel.Body = s.Article.Body
	s.articleModel.Author = GetArticleUserModel(myUserModel)
	s.articleModel.setTags(s.Article.Tags)
	return nil
}

// ArticleUpdateValidator is the schema for PUT /api/articles/:slug. Nullable
// fields give tri-state semantics: absent fields are skipped ("omitnil"),
// null or blank fails "required" on the string fields, and an explicit null
// tagList fails "notnull" (an empty tagList array is a valid value that
// clears all tags).
type ArticleUpdateValidator struct {
	Article struct {
		Title       common.Nullable[string]   `json:"title" binding:"omitnil,required,min=4"`
		Description common.Nullable[string]   `json:"description" binding:"omitnil,required,max=2048"`
		Body        common.Nullable[string]   `json:"body" binding:"omitnil,required,max=2048"`
		TagList     common.Nullable[[]string] `json:"tagList" binding:"omitnil,notnull"`
	} `json:"article"`
}

func NewArticleUpdateValidator() ArticleUpdateValidator {
	return ArticleUpdateValidator{}
}

func (s *ArticleUpdateValidator) Bind(c *gin.Context) error {
	return common.Bind(c, s)
}

// invalidFields lists request fields whose JSON value had the wrong type;
// see UserUpdateValidator.invalidFields in the users package.
func (s *ArticleUpdateValidator) invalidFields() []string {
	fields := []string{}
	if s.Article.Title.IsInvalid() {
		fields = append(fields, "title")
	}
	if s.Article.Description.IsInvalid() {
		fields = append(fields, "description")
	}
	if s.Article.Body.IsInvalid() {
		fields = append(fields, "body")
	}
	if s.Article.TagList.IsInvalid() {
		fields = append(fields, "tagList")
	}
	return fields
}

type CommentModelValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"required,max=2048"`
	} `json:"comment"`
	commentModel CommentModel `json:"-"`
}

func NewCommentModelValidator() CommentModelValidator {
	return CommentModelValidator{}
}

func (s *CommentModelValidator) Bind(c *gin.Context) error {
	myUserModel := c.MustGet("my_user_model").(users.UserModel)

	err := common.Bind(c, s)
	if err != nil {
		return err
	}
	s.commentModel.Body = s.Comment.Body
	s.commentModel.Author = GetArticleUserModel(myUserModel)
	return nil
}
