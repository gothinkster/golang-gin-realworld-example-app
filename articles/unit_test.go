package articles

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var test_db *gorm.DB

func TestArticleModel(t *testing.T) {
	asserts := assert.New(t)

	// Test article creation
	userModel := users.UserModel{
		Username: "testuser",
		Email:    "test@example.com",
		Bio:      "test bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)
	asserts.NotEqual(uint(0), articleUserModel.ID, "ArticleUserModel should be created")
	asserts.Equal(userModel.ID, articleUserModel.UserModelID, "UserModelID should match")

	// Test article creation and save
	article := ArticleModel{
		Slug:        "test-article",
		Title:       "Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	err := SaveOne(&article)
	asserts.NoError(err, "Article should be saved successfully")
	asserts.NotEqual(uint(0), article.ID, "Article ID should be set")

	// Test FindOneArticle
	foundArticle, err := FindOneArticle(&ArticleModel{Slug: "test-article"})
	asserts.NoError(err, "Article should be found")
	asserts.Equal("test-article", foundArticle.Slug, "Slug should match")
	asserts.Equal("Test Article", foundArticle.Title, "Title should match")

	// Test favoritesCount
	count := article.favoritesCount()
	asserts.Equal(uint(0), count, "Favorites count should be 0 initially")

	// Test isFavoriteBy
	isFav := article.isFavoriteBy(articleUserModel)
	asserts.False(isFav, "Article should not be favorited initially")

	// Test favoriteBy
	err = article.favoriteBy(articleUserModel)
	asserts.NoError(err, "Favorite should succeed")

	isFav = article.isFavoriteBy(articleUserModel)
	asserts.True(isFav, "Article should be favorited after favoriteBy")

	count = article.favoritesCount()
	asserts.Equal(uint(1), count, "Favorites count should be 1 after favoriting")

	// Test unFavoriteBy
	err = article.unFavoriteBy(articleUserModel)
	asserts.NoError(err, "UnFavorite should succeed")

	isFav = article.isFavoriteBy(articleUserModel)
	asserts.False(isFav, "Article should not be favorited after unFavoriteBy")

	count = article.favoritesCount()
	asserts.Equal(uint(0), count, "Favorites count should be 0 after unfavoriting")

	// Test article Update
	err = article.Update(map[string]interface{}{"Title": "Updated Title"})
	asserts.NoError(err, "Update should succeed")

	foundArticle, _ = FindOneArticle(&ArticleModel{Slug: article.Slug})
	asserts.Equal("Updated Title", foundArticle.Title, "Title should be updated")

	// Test DeleteArticleModel
	err = DeleteArticleModel(&ArticleModel{Slug: article.Slug})
	asserts.NoError(err, "Delete should succeed")
}

func TestTagModel(t *testing.T) {
	asserts := assert.New(t)

	// Create a tag
	tag := TagModel{Tag: "golang"}
	test_db.Create(&tag)
	asserts.NotEqual(uint(0), tag.ID, "Tag should be created")

	// Test getAllTags
	tags, err := getAllTags()
	asserts.NoError(err, "getAllTags should succeed")
	asserts.GreaterOrEqual(len(tags), 1, "Should have at least one tag")
}

func TestCommentModel(t *testing.T) {
	asserts := assert.New(t)

	// Create user and article
	userModel := users.UserModel{
		Username: "commentuser",
		Email:    "comment@example.com",
		Bio:      "comment bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)

	article := ArticleModel{
		Slug:        "comment-test-article",
		Title:       "Comment Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Create a comment
	comment := CommentModel{
		ArticleID: article.ID,
		AuthorID:  articleUserModel.ID,
		Body:      "Test comment",
	}
	test_db.Create(&comment)
	asserts.NotEqual(uint(0), comment.ID, "Comment should be created")

	// Test getComments
	err := article.getComments()
	asserts.NoError(err, "getComments should succeed")
	asserts.GreaterOrEqual(len(article.Comments), 1, "Should have at least one comment")

	// Test DeleteCommentModel
	err = DeleteCommentModel(&CommentModel{Body: "Test comment"})
	asserts.NoError(err, "DeleteCommentModel should succeed")
}

func TestFindManyArticle(t *testing.T) {
	asserts := assert.New(t)

	// Test FindManyArticle with default params
	articles, count, err := FindManyArticle("", "", "10", "0", "")
	asserts.NoError(err, "FindManyArticle should succeed")
	asserts.GreaterOrEqual(count, 0, "Count should be non-negative")
	asserts.NotNil(articles, "Articles should not be nil")
}

func TestGetArticleFeed(t *testing.T) {
	asserts := assert.New(t)

	// Create a user
	userModel := users.UserModel{
		Username: "feeduser",
		Email:    "feed@example.com",
		Bio:      "feed bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)

	// Test GetArticleFeed
	articles, count, err := articleUserModel.GetArticleFeed("10", "0")
	asserts.NoError(err, "GetArticleFeed should succeed")
	asserts.GreaterOrEqual(count, 0, "Count should be non-negative")
	asserts.NotNil(articles, "Articles should not be nil")
}

func TestSetTags(t *testing.T) {
	asserts := assert.New(t)

	// Create user and article
	userModel := users.UserModel{
		Username: "taguser",
		Email:    "tag@example.com",
		Bio:      "tag bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)

	article := ArticleModel{
		Slug:        "tag-test-article",
		Title:       "Tag Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}

	// Test setTags
	err := article.setTags([]string{"go", "programming", "web"})
	asserts.NoError(err, "setTags should succeed")
	asserts.Equal(3, len(article.Tags), "Should have 3 tags")
}

// Helper functions for router tests
func HeaderTokenMock(req *http.Request, u uint) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", common.GenToken(u)))
}

func userModelMocker(n int) []users.UserModel {
	var offset int64
	test_db.Model(&users.UserModel{}).Count(&offset)
	var ret []users.UserModel
	for i := int(offset) + 1; i <= int(offset)+n; i++ {
		image := fmt.Sprintf("http://image/%v.jpg", i)
		// Generate password hash directly using bcrypt
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Sprintf("failed to generate password hash: %v", err))
		}
		userModel := users.UserModel{
			Username:     fmt.Sprintf("articleuser%v", i),
			Email:        fmt.Sprintf("articleuser%v@test.com", i),
			Bio:          fmt.Sprintf("bio%v", i),
			Image:        &image,
			PasswordHash: string(passwordHash),
		}
		test_db.Create(&userModel)
		ret = append(ret, userModel)
	}
	return ret
}

func resetDBWithMock() {
	common.TestDBFree(test_db)
	test_db = common.TestDBInit()
	users.AutoMigrate()
	test_db.AutoMigrate(&ArticleModel{})
	test_db.AutoMigrate(&TagModel{})
	test_db.AutoMigrate(&FavoriteModel{})
	test_db.AutoMigrate(&ArticleUserModel{})
	test_db.AutoMigrate(&CommentModel{})
	userModelMocker(3)
}

// Router tests
var articleRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexp string
	msg            string
}{
	// Test article list
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/api/articles/",
		"GET",
		``,
		http.StatusOK,
		`{"articles":\[\],"articlesCount":0}`,
		"empty article list should return empty array",
	},
	// Test tags list
	{
		func(req *http.Request) {},
		"/api/tags/",
		"GET",
		``,
		http.StatusOK,
		`{"tags":\[\]}`,
		"empty tags list should return empty array",
	},
	// Test create article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/",
		"POST",
		`{"article":{"title":"Test Article","description":"Test Description","body":"Test Body","tagList":["test","golang"]}}`,
		http.StatusCreated,
		`"title":"Test Article"`,
		"create article should succeed with auth",
	},
	// Test get single article
	{
		func(req *http.Request) {},
		"/api/articles/test-article",
		"GET",
		``,
		http.StatusOK,
		`"slug":"test-article"`,
		"get single article should succeed",
	},
	// Test article list with articles
	{
		func(req *http.Request) {},
		"/api/articles/",
		"GET",
		``,
		http.StatusOK,
		`"articlesCount":1`,
		"article list should contain created article",
	},
	// Test articles by tag
	{
		func(req *http.Request) {},
		"/api/articles/?tag=golang",
		"GET",
		``,
		http.StatusOK,
		`"articles":\[`,
		"articles by tag should work",
	},
	// Test articles by author
	{
		func(req *http.Request) {},
		"/api/articles/?author=articleuser1",
		"GET",
		``,
		http.StatusOK,
		`"articles":\[`,
		"articles by author should work",
	},
	// Test update article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/test-article",
		"PUT",
		`{"article":{"title":"Updated Title"}}`,
		http.StatusOK,
		`"title":"Updated Title"`,
		"update article should succeed",
	},
	// Test favorite article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/updated-title/favorite",
		"POST",
		``,
		http.StatusOK,
		`"favorited":true`,
		"favorite article should succeed",
	},
	// Test favorites count
	{
		func(req *http.Request) {},
		"/api/articles/updated-title",
		"GET",
		``,
		http.StatusOK,
		`"favoritesCount":1`,
		"favorites count should be 1",
	},
	// Test articles favorited by user
	{
		func(req *http.Request) {},
		"/api/articles/?favorited=articleuser1",
		"GET",
		``,
		http.StatusOK,
		`"articlesCount":1`,
		"articles favorited by user should work",
	},
	// Test unfavorite article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/updated-title/favorite",
		"DELETE",
		``,
		http.StatusOK,
		`"favorited":false`,
		"unfavorite article should succeed",
	},
	// Test favorites count after unfavorite
	{
		func(req *http.Request) {},
		"/api/articles/updated-title",
		"GET",
		``,
		http.StatusOK,
		`"favoritesCount":0`,
		"favorites count should be 0 after unfavorite",
	},
	// Test create comment
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/updated-title/comments",
		"POST",
		`{"comment":{"body":"Test comment body"}}`,
		http.StatusCreated,
		`"body":"Test comment body"`,
		"create comment should succeed",
	},
	// Test get comments
	{
		func(req *http.Request) {},
		"/api/articles/updated-title/comments",
		"GET",
		``,
		http.StatusOK,
		`"comments":\[`,
		"get comments should succeed",
	},
	// Test delete comment
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/updated-title/comments/1",
		"DELETE",
		``,
		http.StatusOK,
		``,
		"delete comment should succeed",
	},
	// Test feed (requires auth) - returns empty array since no follow relationship set up
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/api/articles/feed",
		"GET",
		``,
		http.StatusOK,
		`"articles":\[\]`,
		"feed should return empty array when user follows no one",
	},
	// Test delete article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/updated-title",
		"DELETE",
		``,
		http.StatusOK,
		``,
		"delete article should succeed",
	},
	// Test 404 for deleted article
	{
		func(req *http.Request) {},
		"/api/articles/updated-title",
		"GET",
		``,
		http.StatusNotFound,
		`"articles":"Invalid slug"`,
		"deleted article should return 404",
	},
	// Test favorite non-existent article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/non-existent/favorite",
		"POST",
		``,
		http.StatusNotFound,
		`"articles":"Invalid slug"`,
		"favorite non-existent article should return 404",
	},
	// Test unfavorite non-existent article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/non-existent/favorite",
		"DELETE",
		``,
		http.StatusNotFound,
		`"articles":"Invalid slug"`,
		"unfavorite non-existent article should return 404",
	},
	// Test create article with invalid data
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/",
		"POST",
		`{"article":{"title":"ab","description":"Test","body":"Test"}}`,
		http.StatusUnprocessableEntity,
		`"errors"`,
		"create article with short title should fail",
	},
	// Test create comment on non-existent article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/non-existent/comments",
		"POST",
		`{"comment":{"body":"Test"}}`,
		http.StatusNotFound,
		`"comment":"Invalid slug"`,
		"create comment on non-existent article should return 404",
	},
	// Test get comments on non-existent article
	{
		func(req *http.Request) {},
		"/api/articles/non-existent/comments",
		"GET",
		``,
		http.StatusNotFound,
		`"comments":"Invalid slug"`,
		"get comments on non-existent article should return 404",
	},
	// Test update non-existent article
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/non-existent",
		"PUT",
		`{"article":{"title":"Test"}}`,
		http.StatusNotFound,
		`"articles":"Invalid slug"`,
		"update non-existent article should return 404",
	},
	// Test delete non-existent article (GORM delete returns OK even if not found)
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/non-existent",
		"DELETE",
		``,
		http.StatusOK,
		``,
		"delete non-existent article returns OK (soft delete behavior)",
	},
	// Test delete comment with invalid id
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/api/articles/test/comments/invalid",
		"DELETE",
		``,
		http.StatusNotFound,
		`"comment":"Invalid id"`,
		"delete comment with invalid id should return 404",
	},
}

func TestArticleRouters(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(users.AuthMiddleware(false))
	ArticlesAnonymousRegister(r.Group("/api/articles"))
	TagsAnonymousRegister(r.Group("/api/tags"))
	r.Use(users.AuthMiddleware(true))
	ArticlesRegister(r.Group("/api/articles"))

	for _, testData := range articleRequestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest(testData.method, testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		testData.init(req)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		if testData.responseRegexp != "" {
			asserts.Regexp(testData.responseRegexp, w.Body.String(), "Response Content - "+testData.msg)
		}
	}
}

// This is a hack way to add test database for each case
func TestMain(m *testing.M) {
	test_db = common.TestDBInit()
	users.AutoMigrate()
	test_db.AutoMigrate(&ArticleModel{})
	test_db.AutoMigrate(&TagModel{})
	test_db.AutoMigrate(&FavoriteModel{})
	test_db.AutoMigrate(&ArticleUserModel{})
	test_db.AutoMigrate(&CommentModel{})
	exitVal := m.Run()
	common.TestDBFree(test_db)
	os.Exit(exitVal)
}
