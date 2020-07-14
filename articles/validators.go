package articles

import (
	"fmt"
	"golang-gin-realworld-example-app/common"
	"golang-gin-realworld-example-app/users"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type ArticleModelValidator struct {
	Article struct {
		Title       string   `form:"title" json:"title" binding:"required"`
		Description string   `form:"description" json:"description" binding:"max=2048"`
		Body        string   `form:"body" json:"body" binding:"max=2048"`
		Tags        []string `form:"tagList" json:"tagList"`
	} `json:"article"`
	articleModel ArticleModel `json:"-"`
}

type CommentVoteValidator struct {
	CommentID uint `uri:"id" binding:"required"`
	UpVote    bool `json:"up_vote"`
	DownVote  bool `json:"down_vote"`
}

type CommentVoteError struct {
	err string
}

func (e CommentVoteError) Error() string {
	return e.err
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

type CommentModelValidator struct {
	Comment struct {
		Body string `form:"body" json:"body" binding:"max=2048"`
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

func (s *CommentVoteValidator) Bind(c *gin.Context) error {
	// commentID := c.Param("id")(int)
	// s.CommentID = commentID
	if err := c.ShouldBindUri(s); err != nil {
		return err
	}
	err := common.Bind(c, s)
	if err != nil {
		return err
	}
	fmt.Printf("\n\n In Bind: %+v \n\n", s)
	if (s.DownVote && s.UpVote) || !(s.UpVote || s.DownVote) {
		err := CommentVoteError{err: "Either one of the UpVote or DownVote can be true not both."}
		return err
	}
	return nil
}

func (s *CommentVoteValidator) BindCommentId(c *gin.Context) error {
	if err := c.ShouldBindUri(s); err != nil {
		return err
	}
	fmt.Printf("\n\n In Bind: %+v \n\n", s)
	return nil
}
