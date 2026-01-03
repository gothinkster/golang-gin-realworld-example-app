package articles

import (
	"os"
	"testing"

	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gothinkster/golang-gin-realworld-example-app/users"
	"github.com/stretchr/testify/assert"
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
