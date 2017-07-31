package articles

import (
    "github.com/gosimple/slug"
    "gopkg.in/gin-gonic/gin.v1"
    "golang-gin-starter-kit/users"
)

type ArticleSerializer struct {
    C *gin.Context
    ArticleModel
}


type TagSerializer struct {
    TagModel
}

func (self *TagSerializer) Response() string {
    return self.TagModel.Tag
}

type ArticleResponse struct {
    ID              uint        `json:"-"`
    Title           string      `json:"title"`
    Slug            string      `json:"slug"`
    Description     string      `json:"description"`
    Body            string      `json:"body"`
    CreatedAt       string      `json:"createdAt"`
    UpdatedAt       string      `json:"updatedAt"`
    Author          users.ProfileResponse `json:"author"`
    Tags            []string    `json:"tagList"`
}

func (self *ArticleSerializer) Response() ArticleResponse {
    authorSerializer := users.ProfileSerializer{self.C,self.Author}
    article := ArticleResponse{
        ID:             self.ID,
        Slug:           slug.Make(self.Title),
        Title:          self.Title,
        Description:    self.Description,
        Body:           self.Body,
        CreatedAt:      self.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        //UpdatedAt:      self.UpdatedAt.UTC().Format(time.RFC3339Nano),
        UpdatedAt:      self.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        Author:         authorSerializer.Response(),
    }
    for _, tag := range self.Tags {
        tagSerializer := TagSerializer{tag}
        article.Tags = append(article.Tags, tagSerializer.Response())
    }
    return article
}

