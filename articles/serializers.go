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

type ArticleResponse struct {
    ID              uint        `json:"-"`
    Title           string      `json:"title"`
    Slug            string      `json:"slug"`
    Description     string      `json:"description"`
    Body            string      `json:"body"`
    CreatedAt       string      `json:"createdAt"`
    UpdatedAt       string      `json:"updatedAt"`
    Author          users.ProfileResponse `json:"author"`
}

func (self *ArticleSerializer) Response() ArticleResponse {
    author := users.ProfileSerializer{self.C,self.Author}
    article := ArticleResponse{
        ID:             self.ID,
        Slug:           slug.Make(self.Title),
        Title:          self.Title,
        Description:    self.Description,
        Body:           self.Body,
        CreatedAt:      self.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        UpdatedAt:      self.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
        Author:         author.Response(),
        //UpdatedAt:      self.UpdatedAt.UTC().Format(time.RFC3339Nano),
    }
    return article
}

