package articles

import (
    "github.com/gosimple/slug"
    "gopkg.in/gin-gonic/gin.v1"
    "golang-gin-starter-kit/users"
)

type TagSerializer struct {
    C *gin.Context
    TagModel
}

type TagsSerializer struct {
    C    *gin.Context
    Tags []TagModel
}

func (self *TagSerializer) Response() string {
    return self.TagModel.Tag
}

func (self *TagsSerializer) Response() []string {
    response := []string{}
    for _, tag := range self.Tags {
        serializer := TagSerializer{self.C, tag}
        response = append(response, serializer.Response())
    }
    return response
}

type ArticleUserSerializer struct {
    C *gin.Context
    ArticleUserModel
}

func (self *ArticleUserSerializer) Response() users.ProfileResponse {
    response := users.ProfileSerializer{self.C, self.ArticleUserModel.UserModel}
    return response.Response()
}

type ArticleSerializer struct {
    C *gin.Context
    ArticleModel
}

type ArticleResponse struct {
    ID             uint        `json:"-"`
    Title          string      `json:"title"`
    Slug           string      `json:"slug"`
    Description    string      `json:"description"`
    Body           string      `json:"body"`
    CreatedAt      string      `json:"createdAt"`
    UpdatedAt      string      `json:"updatedAt"`
    Author         users.ProfileResponse `json:"author"`
    Tags           []string    `json:"tagList"`
    Favorite       bool        `json:"favorited"`
    FavoritesCount uint        `json:"favoritesCount"`
}

type ArticlesSerializer struct {
    C        *gin.Context
    Articles []ArticleModel
}

func (self *ArticleSerializer) Response() ArticleResponse {
    myUserModel := self.C.MustGet("my_user_model").(users.UserModel)
    authorSerializer := ArticleUserSerializer{self.C, self.Author}
    response := ArticleResponse{
        ID:          self.ID,
        Slug:        slug.Make(self.Title),
        Title:       self.Title,
        Description: self.Description,
        Body:        self.Body,
        CreatedAt:   self.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        //UpdatedAt:      self.UpdatedAt.UTC().Format(time.RFC3339Nano),
        UpdatedAt:      self.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        Author:         authorSerializer.Response(),
        Favorite:       self.isFavoriteBy(GetArticleUserModel(myUserModel)),
        FavoritesCount: self.favoritesCount(),
    }
    response.Tags = make([]string, 0)
    for _, tag := range self.Tags {
        serializer := TagSerializer{self.C, tag}
        response.Tags = append(response.Tags, serializer.Response())
    }
    return response
}

func (self *ArticlesSerializer) Response() []ArticleResponse {
    response := []ArticleResponse{}
    for _, article := range self.Articles {
        serializer := ArticleSerializer{self.C, article}
        response = append(response, serializer.Response())
    }
    return response
}

type CommentSerializer struct {
    C *gin.Context
    CommentModel
}

type CommentsSerializer struct {
    C        *gin.Context
    Comments []CommentModel
}

type CommentResponse struct {
    ID        uint        `json:"id"`
    Body      string      `json:"body"`
    CreatedAt string      `json:"createdAt"`
    UpdatedAt string      `json:"updatedAt"`
    Author    users.ProfileResponse `json:"author"`
}

func (self *CommentSerializer) Response() CommentResponse {
    authorSerializer := ArticleUserSerializer{self.C, self.Author}
    response := CommentResponse{
        ID:        self.ID,
        Body:      self.Body,
        CreatedAt: self.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        UpdatedAt: self.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        Author:    authorSerializer.Response(),
    }
    return response
}

func (self *CommentsSerializer) Response() []CommentResponse {
    response := []CommentResponse{}
    for _, comment := range self.Comments {
        serializer := CommentSerializer{self.C, comment}
        response = append(response, serializer.Response())
    }
    return response
}
