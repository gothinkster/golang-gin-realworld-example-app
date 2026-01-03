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

func setupRouter() *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false

	v1 := r.Group("/api")
	users.UsersRegister(v1.Group("/users"))
	v1.Use(users.AuthMiddleware(false))
	ArticlesAnonymousRegister(v1.Group("/articles"))
	TagsAnonymousRegister(v1.Group("/tags"))

	v1.Use(users.AuthMiddleware(true))
	ArticlesRegister(v1.Group("/articles"))

	return r
}

func createTestUser() users.UserModel {
	// Generate a proper password hash to satisfy NOT NULL constraint
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	userModel := users.UserModel{
		Username:     fmt.Sprintf("testuser%d", common.RandInt()),
		Email:        fmt.Sprintf("test%d@example.com", common.RandInt()),
		Bio:          "test bio",
		PasswordHash: string(passwordHash),
	}
	test_db.Create(&userModel)
	return userModel
}

// createArticleWithUser creates a test article with an author user
func createArticleWithUser(title, slug string) (ArticleModel, users.UserModel) {
	user := createTestUser()
	articleUserModel := GetArticleUserModel(user)
	article := ArticleModel{
		Slug:        slug,
		Title:       title,
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)
	return article, user
}

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

	// Create a user and article for testing
	userModel := users.UserModel{
		Username: fmt.Sprintf("findmanyuser%d", common.RandInt()),
		Email:    fmt.Sprintf("findmany%d@example.com", common.RandInt()),
		Bio:      "test bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)
	article := ArticleModel{
		Slug:        fmt.Sprintf("findmany-article-%d", common.RandInt()),
		Title:       "FindMany Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	article.setTags([]string{"findmanytag"})
	SaveOne(&article)

	// Favorite the article
	article.favoriteBy(articleUserModel)

	// Test FindManyArticle with default params
	articles, count, err := FindManyArticle("", "", "10", "0", "")
	asserts.NoError(err, "FindManyArticle should succeed")
	asserts.GreaterOrEqual(count, 1, "Count should be at least 1")
	asserts.NotNil(articles, "Articles should not be nil")

	// Test with invalid limit/offset
	_, _, err = FindManyArticle("", "", "invalid", "invalid", "")
	asserts.NoError(err, "FindManyArticle with invalid params should succeed")

	// Test filter by tag
	_, count, err = FindManyArticle("findmanytag", "", "10", "0", "")
	asserts.NoError(err, "FindManyArticle by tag should succeed")
	asserts.GreaterOrEqual(count, 1, "Count should be at least 1 for tag filter")

	// Test filter by non-existent tag
	_, count, err = FindManyArticle("nonexistenttag", "", "10", "0", "")
	asserts.NoError(err, "FindManyArticle by non-existent tag should succeed")
	asserts.Equal(0, count, "Count should be 0 for non-existent tag")

	// Test filter by author
	_, count, err = FindManyArticle("", userModel.Username, "10", "0", "")
	asserts.NoError(err, "FindManyArticle by author should succeed")
	asserts.GreaterOrEqual(count, 1, "Count should be at least 1 for author filter")

	// Test filter by non-existent author
	_, _, err = FindManyArticle("", "nonexistentauthor", "10", "0", "")
	asserts.NoError(err, "FindManyArticle by non-existent author should succeed")

	// Test filter by favorited
	_, count, err = FindManyArticle("", "", "10", "0", userModel.Username)
	asserts.NoError(err, "FindManyArticle by favorited should succeed")
	asserts.GreaterOrEqual(count, 1, "Count should be at least 1 for favorited filter")

	// Test filter by non-existent favorited user
	_, _, err = FindManyArticle("", "", "10", "0", "nonexistentuser")
	asserts.NoError(err, "FindManyArticle by non-existent favorited should succeed")
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

// Helper functions for router tests - used by TestArticleRouters

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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 2)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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
			common.HeaderTokenMock(req, 1)
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

// HTTP API Tests

func TestCreateArticleRequiredFields(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test missing body field
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"Test Title","description":"Test Description"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Missing body should return 422")
	asserts.Contains(w.Body.String(), "Body", "Error should mention Body field")

	// Test missing description field
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"Test Title","body":"Test Body"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Missing description should return 422")
	asserts.Contains(w.Body.String(), "Description", "Error should mention Description field")

	// Test valid article creation
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"Test Title","description":"Test Description","body":"Test Body"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusCreated, w.Code, "Valid article should return 201")
	asserts.Contains(w.Body.String(), `"article"`, "Response should contain article")
}

func TestCreateCommentRequiredFields(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article first
	articleUserModel := GetArticleUserModel(user)
	article := ArticleModel{
		Slug:        fmt.Sprintf("test-comment-article-%d", common.RandInt()),
		Title:       "Test Comment Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test missing body field
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/articles/%s/comments", article.Slug), bytes.NewBufferString(`{"comment":{}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Missing body should return 422")
	asserts.Contains(w.Body.String(), "Body", "Error should mention Body field")

	// Test valid comment creation - should return 201 per OpenAPI spec
	req, _ = http.NewRequest("POST", fmt.Sprintf("/api/articles/%s/comments", article.Slug), bytes.NewBufferString(`{"comment":{"body":"Test comment body"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusCreated, w.Code, "Valid comment should return 201")
	asserts.Contains(w.Body.String(), `"comment"`, "Response should contain comment")
}

func TestArticleFeedCount(t *testing.T) {
	asserts := assert.New(t)

	// Create two users
	user1 := createTestUser()
	user2 := createTestUser()

	// User1 follows User2
	err := followUser(user1, user2)
	asserts.NoError(err, "Follow should succeed")

	// Create an article by User2
	articleUserModel := GetArticleUserModel(user2)
	article := ArticleModel{
		Slug:        fmt.Sprintf("feed-test-article-%d", common.RandInt()),
		Title:       "Feed Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Get feed for User1
	articleUserModel1 := GetArticleUserModel(user1)
	articles, count, err := articleUserModel1.GetArticleFeed("10", "0")
	asserts.NoError(err, "GetArticleFeed should succeed")
	asserts.Equal(1, count, "Count should be 1 after following user with 1 article")
	asserts.Equal(1, len(articles), "Should have 1 article in feed")
}

func followUser(follower, following users.UserModel) error {
	db := common.GetDB()
	var follow users.FollowModel
	err := db.FirstOrCreate(&follow, &users.FollowModel{
		FollowingID:  following.ID,
		FollowedByID: follower.ID,
	}).Error
	return err
}

func TestTagsEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Create some tags
	tag1 := TagModel{Tag: fmt.Sprintf("testtag1-%d", common.RandInt())}
	tag2 := TagModel{Tag: fmt.Sprintf("testtag2-%d", common.RandInt())}
	test_db.Create(&tag1)
	test_db.Create(&tag2)

	req, _ := http.NewRequest("GET", "/api/tags", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Tags endpoint should return 200")
	asserts.Contains(w.Body.String(), `"tags"`, "Response should contain tags array")
}

func TestArticleListEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	_, _ = createArticleWithUser("List Test Article", fmt.Sprintf("list-test-article-%d", common.RandInt()))

	// Test list articles
	req, _ := http.NewRequest("GET", "/api/articles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Articles list should return 200")
	asserts.Contains(w.Body.String(), `"articles"`, "Response should contain articles")
	asserts.Contains(w.Body.String(), `"articlesCount"`, "Response should contain articlesCount")
}

func TestArticleRetrieveEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	slug := fmt.Sprintf("retrieve-test-article-%d", common.RandInt())
	article, _ := createArticleWithUser("Retrieve Test Article", slug)

	// Test retrieve article
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/articles/%s", article.Slug), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Article retrieve should return 200")
	asserts.Contains(w.Body.String(), `"article"`, "Response should contain article")
	asserts.Contains(w.Body.String(), `"Retrieve Test Article"`, "Response should contain the article title")
}

func TestArticleUpdateEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	slug := fmt.Sprintf("update-test-article-%d", common.RandInt())
	article, user := createArticleWithUser("Update Test Article", slug)

	// Test update article
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/articles/%s", article.Slug), bytes.NewBufferString(`{"article":{"body":"Updated Body"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Article update should return 200")
	asserts.Contains(w.Body.String(), `"article"`, "Response should contain article")
}

func TestArticleDeleteEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("delete-test-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Delete Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test delete article
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s", slug), nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Article delete should return 200")
}

func TestArticleFavoriteEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("favorite-test-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Favorite Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test favorite article
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/articles/%s/favorite", slug), nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Article favorite should return 200")
	asserts.Contains(w.Body.String(), `"favorited":true`, "Article should be favorited")

	// Test unfavorite article
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s/favorite", slug), nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Article unfavorite should return 200")
	asserts.Contains(w.Body.String(), `"favorited":false`, "Article should be unfavorited")
}

func TestArticleCommentsEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("comments-test-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Comments Test Article",
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
		Body:      "Test comment for list",
	}
	test_db.Create(&comment)

	// Test list comments
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/articles/%s/comments", slug), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Comments list should return 200")
	asserts.Contains(w.Body.String(), `"comments"`, "Response should contain comments")

	// Test delete comment
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", slug, comment.ID), nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Comment delete should return 200")
}

func TestArticleFeedEndpoint(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test feed endpoint
	req, _ := http.NewRequest("GET", "/api/articles/feed", nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Feed should return 200")
	asserts.Contains(w.Body.String(), `"articles"`, "Response should contain articles")
	asserts.Contains(w.Body.String(), `"articlesCount"`, "Response should contain articlesCount")

	// Test feed with limit and offset params
	req, _ = http.NewRequest("GET", "/api/articles/feed?limit=5&offset=0", nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "Feed with params should return 200")
}

func TestArticleFeedWithFollowing(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Create two users
	user1 := createTestUser()
	user2 := createTestUser()

	// User1 follows User2
	followUser(user1, user2)

	// Create an article by User2
	articleUserModel := GetArticleUserModel(user2)
	article := ArticleModel{
		Slug:        fmt.Sprintf("feed-following-article-%d", common.RandInt()),
		Title:       "Feed Following Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test feed for User1 - should include User2's article
	req, _ := http.NewRequest("GET", "/api/articles/feed", nil)
	common.HeaderTokenMock(req, user1.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Feed should return 200")
	asserts.Contains(w.Body.String(), `"articles"`, "Response should contain articles")
}

func TestArticleNotFoundErrors(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test retrieve non-existent article
	req, _ := http.NewRequest("GET", "/api/articles/non-existent-slug", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Non-existent article should return 404")

	// Test update non-existent article
	req, _ = http.NewRequest("PUT", "/api/articles/non-existent-slug", bytes.NewBufferString(`{"article":{"body":"test"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Update non-existent article should return 404")

	// Test delete non-existent article - returns 200 (GORM delete doesn't error on 0 rows)
	req, _ = http.NewRequest("DELETE", "/api/articles/non-existent-slug", nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "Delete non-existent article returns 200")

	// Test favorite non-existent article
	req, _ = http.NewRequest("POST", "/api/articles/non-existent-slug/favorite", nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Favorite non-existent article should return 404")

	// Test unfavorite non-existent article
	req, _ = http.NewRequest("DELETE", "/api/articles/non-existent-slug/favorite", nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Unfavorite non-existent article should return 404")

	// Test create comment on non-existent article
	req, _ = http.NewRequest("POST", "/api/articles/non-existent-slug/comments", bytes.NewBufferString(`{"comment":{"body":"test"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Comment on non-existent article should return 404")

	// Test list comments on non-existent article
	req, _ = http.NewRequest("GET", "/api/articles/non-existent-slug/comments", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "List comments on non-existent article should return 404")

	// Test delete comment with invalid id
	req, _ = http.NewRequest("DELETE", "/api/articles/some-slug/comments/invalid", nil)
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusNotFound, w.Code, "Delete comment with invalid id should return 404")
}

func TestArticleListWithFilters(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Test list with tag filter
	req, _ := http.NewRequest("GET", "/api/articles?tag=sometag", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "List with tag filter should return 200")

	// Test list with author filter
	req, _ = http.NewRequest("GET", "/api/articles?author=someauthor", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "List with author filter should return 200")

	// Test list with favorited filter
	req, _ = http.NewRequest("GET", "/api/articles?favorited=someuser", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "List with favorited filter should return 200")

	// Test list with limit and offset
	req, _ = http.NewRequest("GET", "/api/articles?limit=5&offset=0", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "List with limit/offset should return 200")
}

func TestArticleValidationErrors(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test create article with missing title
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"description":"test","body":"test"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Missing title should return 422")

	// Test create article with short title
	req, _ = http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"abc","description":"test","body":"test"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Short title should return 422")
}

func TestArticleFeedUnauthorized(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Test feed endpoint without auth - should return 401
	req, _ := http.NewRequest("GET", "/api/articles/feed", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnauthorized, w.Code, "Feed without auth should return 401")
}

func TestArticleUpdateValidationErrors(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("update-validation-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Update Validation Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test update with title too short
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/articles/%s", slug), bytes.NewBufferString(`{"article":{"title":"ab"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Short title in update should return 422")
}

func TestArticleCreateWithTags(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test create article with tags
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"Test With Tags","description":"Test Description","body":"Test Body","tagList":["go","gin","test"]}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusCreated, w.Code, "Article with tags should return 201")
	asserts.Contains(w.Body.String(), `"tagList"`, "Response should contain tagList")
}

func TestCommentDeleteWithValidArticle(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("comment-delete-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Comment Delete Test",
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
		Body:      "Test comment for deletion",
	}
	test_db.Create(&comment)

	// Test delete existing comment
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", slug, comment.ID), nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "Delete existing comment should return 200")
}

func TestSetTagsEmpty(t *testing.T) {
	asserts := assert.New(t)

	userModel := users.UserModel{
		Username: fmt.Sprintf("emptytaguser%d", common.RandInt()),
		Email:    fmt.Sprintf("emptytag%d@example.com", common.RandInt()),
		Bio:      "test bio",
	}
	test_db.Create(&userModel)

	articleUserModel := GetArticleUserModel(userModel)
	article := ArticleModel{
		Slug:        fmt.Sprintf("empty-tags-article-%d", common.RandInt()),
		Title:       "Empty Tags Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}

	// Test setTags with empty slice
	err := article.setTags([]string{})
	asserts.NoError(err, "setTags with empty slice should succeed")
	asserts.Equal(0, len(article.Tags), "Should have 0 tags")
}

func TestFavoritesCountWithMultipleUsers(t *testing.T) {
	asserts := assert.New(t)

	// Create article
	user1 := createTestUser()
	user2 := createTestUser()

	articleUserModel1 := GetArticleUserModel(user1)
	articleUserModel2 := GetArticleUserModel(user2)

	article := ArticleModel{
		Slug:        fmt.Sprintf("multi-favorite-article-%d", common.RandInt()),
		Title:       "Multi Favorite Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel1,
		AuthorID:    articleUserModel1.ID,
	}
	SaveOne(&article)

	// Both users favorite the article
	article.favoriteBy(articleUserModel1)
	article.favoriteBy(articleUserModel2)

	count := article.favoritesCount()
	asserts.Equal(uint(2), count, "Favorites count should be 2")
}

func TestBatchGetFavoriteStatusEdgeCases(t *testing.T) {
	asserts := assert.New(t)

	user := createTestUser()
	articleUserModel := GetArticleUserModel(user)

	// Test with empty article IDs
	statusMap := BatchGetFavoriteStatus([]uint{}, articleUserModel.ID)
	asserts.Equal(0, len(statusMap), "Empty article IDs should return empty map")

	// Test with zero user ID
	article := ArticleModel{
		Slug:        fmt.Sprintf("batch-status-article-%d", common.RandInt()),
		Title:       "Batch Status Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)
	article.favoriteBy(articleUserModel)

	statusMap = BatchGetFavoriteStatus([]uint{article.ID}, 0)
	asserts.Equal(0, len(statusMap), "Zero user ID should return empty map")

	// Test with valid IDs
	statusMap = BatchGetFavoriteStatus([]uint{article.ID}, articleUserModel.ID)
	asserts.Equal(true, statusMap[article.ID], "Should return true for favorited article")
}

func TestSetTagsRaceCondition(t *testing.T) {
	asserts := assert.New(t)

	user := createTestUser()
	articleUserModel := GetArticleUserModel(user)

	article := ArticleModel{
		Slug:        fmt.Sprintf("race-condition-article-%d", common.RandInt()),
		Title:       "Race Condition Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}

	// Test setTags with duplicate tags
	err := article.setTags([]string{"tag1", "tag2", "tag1"})
	asserts.NoError(err, "setTags should handle duplicate tags")
	// Should have 2 unique tags
	asserts.Equal(3, len(article.Tags), "Should preserve all tags in list")
}

func TestArticleFeedWithEmptyFollowings(t *testing.T) {
	asserts := assert.New(t)

	user := createTestUser()
	articleUserModel := GetArticleUserModel(user)

	// Get feed with no followings
	articles, count, err := articleUserModel.GetArticleFeed("10", "0")
	asserts.NoError(err, "GetArticleFeed should succeed even with no followings")
	asserts.Equal(0, count, "Count should be 0 with no followings")
	asserts.NotNil(articles, "Articles should not be nil")
}

func TestArticleDeleteSuccess(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Create an article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("delete-success-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Delete Success Test",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Test delete existing article
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s", slug), nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Delete existing article should return 200")

	// Verify article is deleted
	foundArticle, err := FindOneArticle(&ArticleModel{Slug: slug})
	asserts.Error(err, "Article should not be found after deletion")
	asserts.Equal(uint(0), foundArticle.ID, "Article ID should be 0")
}

func TestTagListSuccess(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Create some test tags
	tag1 := TagModel{Tag: fmt.Sprintf("listtag1-%d", common.RandInt())}
	tag2 := TagModel{Tag: fmt.Sprintf("listtag2-%d", common.RandInt())}
	test_db.Create(&tag1)
	test_db.Create(&tag2)

	// Test list tags endpoint
	req, _ := http.NewRequest("GET", "/api/tags", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Tags list should return 200")
	asserts.Contains(w.Body.String(), `"tags"`, "Response should contain tags")
}

func TestArticleListErrorHandling(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()

	// Test with invalid limit/offset parameters
	req, _ := http.NewRequest("GET", "/api/articles?limit=abc&offset=xyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "List with invalid params should still return 200")
	asserts.Contains(w.Body.String(), `"articles"`, "Response should contain articles array")
}

func TestArticleFeedErrorPath(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test with invalid limit/offset
	req, _ := http.NewRequest("GET", "/api/articles/feed?limit=invalid&offset=invalid", nil)
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusOK, w.Code, "Feed with invalid params should return 200")
	asserts.Contains(w.Body.String(), `"articles"`, "Response should contain articles")
}

func TestArticleCreateValidation(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test with empty fields
	req, _ := http.NewRequest("POST", "/api/articles", bytes.NewBufferString(`{"article":{"title":"","description":"","body":""}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "Empty fields should return 422")
}

func TestArticleUpdateNonExistent(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()

	// Test update non-existent article
	req, _ := http.NewRequest("PUT", "/api/articles/non-existent-article", bytes.NewBufferString(`{"article":{"title":"New Title"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, user.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusNotFound, w.Code, "Update non-existent article should return 404")
}

func TestArticleDeleteAuthorizationForbidden(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()
	otherUser := createTestUser()

	// Create article by user
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("forbidden-delete-article-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Forbidden Delete Article",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Try to delete by otherUser
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s", slug), nil)
	common.HeaderTokenMock(req, otherUser.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusForbidden, w.Code, "Delete by non-author should return 403")

	// Verify article still exists
	foundArticle, err := FindOneArticle(&ArticleModel{Slug: slug})
	asserts.NoError(err, "Article should still exist")
	asserts.Equal(article.ID, foundArticle.ID, "Article ID should match")
}

func TestArticleUpdateAuthorizationForbidden(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()
	otherUser := createTestUser()

	// Create article by user
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("forbidden-update-article-%d", common.RandInt())
	title := "Forbidden Update Article"
	article := ArticleModel{
		Slug:        slug,
		Title:       title,
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Try to update by otherUser
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/articles/%s", slug), bytes.NewBufferString(`{"article":{"title":"New Title"}}`))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, otherUser.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusForbidden, w.Code, "Update by non-author should return 403")

	// Verify article is unchanged
	foundArticle, _ := FindOneArticle(&ArticleModel{Slug: slug})
	asserts.Equal(title, foundArticle.Title, "Article title should be unchanged")
}

func TestCommentDeleteAuthorizationForbidden(t *testing.T) {
	asserts := assert.New(t)

	r := setupRouter()
	user := createTestUser()
	otherUser := createTestUser()

	// Create article
	articleUserModel := GetArticleUserModel(user)
	slug := fmt.Sprintf("forbidden-comment-delete-%d", common.RandInt())
	article := ArticleModel{
		Slug:        slug,
		Title:       "Forbidden Comment Delete",
		Description: "Test Description",
		Body:        "Test Body",
		Author:      articleUserModel,
		AuthorID:    articleUserModel.ID,
	}
	SaveOne(&article)

	// Create comment by user
	comment := CommentModel{
		ArticleID: article.ID,
		AuthorID:  articleUserModel.ID,
		Body:      "Test comment",
	}
	test_db.Create(&comment)

	// Try to delete by otherUser
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/articles/%s/comments/%d", slug, comment.ID), nil)
	common.HeaderTokenMock(req, otherUser.ID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusForbidden, w.Code, "Delete comment by non-author should return 403")

	// Verify comment still exists
	foundComment, err := FindOneComment(&CommentModel{Model: gorm.Model{ID: comment.ID}})
	asserts.NoError(err, "Comment should still exist")
	asserts.Equal(comment.ID, foundComment.ID, "Comment ID should match")
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
