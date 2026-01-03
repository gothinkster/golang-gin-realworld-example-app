package common

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestConnectingDatabase(t *testing.T) {
	asserts := assert.New(t)
	db := Init()
	dbPath := GetDBPath()
	// Test create & close DB
	_, err := os.Stat(dbPath)
	asserts.NoError(err, "Db should exist")
	sqlDB, _ := db.DB()
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")

	// Test get a connecting from connection pools
	connection := GetDB()
	sqlDB, _ = connection.DB()
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")
	sqlDB.Close()

	// Test DB exceptions
	os.Chmod(dbPath, 0000)
	db = Init()
	sqlDB, _ = db.DB()
	asserts.Error(sqlDB.Ping(), "Db should not be able to ping")
	sqlDB.Close()
	os.Chmod(dbPath, 0644)
}

func TestConnectingTestDatabase(t *testing.T) {
	asserts := assert.New(t)
	// Test create & close DB
	db := TestDBInit()
	testDBPath := GetTestDBPath()
	_, err := os.Stat(testDBPath)
	asserts.NoError(err, "Db should exist")
	sqlDB, _ := db.DB()
	asserts.NoError(sqlDB.Ping(), "Db should be able to ping")
	TestDBFree(db)

	// Test close delete DB
	db = TestDBInit()
	TestDBFree(db)
	_, err = os.Stat(testDBPath)

	asserts.Error(err, "Db should not exist")
}

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	str := RandString(0)
	asserts.Equal(str, "", "length should be ''")

	str = RandString(10)
	asserts.Equal(len(str), 10, "length should be 10")
	for _, ch := range str {
		asserts.Contains(letters, ch, "char should be a-z|A-Z|0-9")
	}
}

func TestRandInt(t *testing.T) {
	asserts := assert.New(t)

	// Test that RandInt returns a value in valid range
	val := RandInt()
	asserts.GreaterOrEqual(val, 0, "RandInt should be >= 0")
	asserts.Less(val, 1000000, "RandInt should be < 1000000")

	// Test multiple calls return different values (statistically)
	vals := make(map[int]bool)
	for i := 0; i < 10; i++ {
		vals[RandInt()] = true
	}
	asserts.Greater(len(vals), 1, "RandInt should return varied values")
}

func TestGenToken(t *testing.T) {
	asserts := assert.New(t)

	token := GenToken(2)

	asserts.IsType(token, string("token"), "token type should be string")
	asserts.Len(token, 115, "JWT's length should be 115")
}

func TestGenTokenMultipleUsers(t *testing.T) {
	asserts := assert.New(t)

	token1 := GenToken(1)
	token2 := GenToken(2)
	token100 := GenToken(100)

	asserts.NotEqual(token1, token2, "Different user IDs should generate different tokens")
	asserts.NotEqual(token2, token100, "Different user IDs should generate different tokens")
	// Token length can vary by 1 character due to timestamp changes
	asserts.GreaterOrEqual(len(token1), 114, "JWT's length should be >= 114 for user 1")
	asserts.LessOrEqual(len(token1), 120, "JWT's length should be <= 120 for user 1")
	asserts.GreaterOrEqual(len(token100), 114, "JWT's length should be >= 114 for user 100")
	asserts.LessOrEqual(len(token100), 120, "JWT's length should be <= 120 for user 100")
}

func TestNewValidatorError(t *testing.T) {
	asserts := assert.New(t)

	type Login struct {
		Username string `form:"username" json:"username" binding:"required,alphanum,min=4,max=255"`
		Password string `form:"password" json:"password" binding:"required,min=8,max=255"`
	}

	var requestTests = []struct {
		bodyData       string
		expectedCode   int
		responseRegexg string
		msg            string
	}{
		{
			`{"username": "wangzitian0","password": "0123456789"}`,
			http.StatusOK,
			`{"status":"you are logged in"}`,
			"valid data and should return StatusCreated",
		},
		{
			`{"username": "wangzitian0","password": "01234567866"}`,
			http.StatusUnauthorized,
			`{"errors":{"user":"wrong username or password"}}`,
			"wrong login status should return StatusUnauthorized",
		},
		{
			`{"username": "wangzitian0","password": "0122"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Password":"{min: 8}"}}`,
			"invalid password of too short and should return StatusUnprocessableEntity",
		},
		{
			`{"username": "_wangzitian0","password": "0123456789"}`,
			http.StatusUnprocessableEntity,
			`{"errors":{"Username":"{key: alphanum}"}}`,
			"invalid username of non alphanum and should return StatusUnprocessableEntity",
		},
	}

	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var json Login
		if err := Bind(c, &json); err == nil {
			if json.Username == "wangzitian0" && json.Password == "0123456789" {
				c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
			} else {
				c.JSON(http.StatusUnauthorized, NewError("user", errors.New("wrong username or password")))
			}
		} else {
			c.JSON(http.StatusUnprocessableEntity, NewValidatorError(err))
		}
	})

	for _, testData := range requestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest("POST", "/login", bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

func TestNewError(t *testing.T) {
	assert := assert.New(t)

	db := TestDBInit()
	type NotExist struct {
		heheda string
	}
	db.AutoMigrate(NotExist{})

	commenError := NewError("database", db.Find(NotExist{heheda: "heheda"}).Error)
	assert.IsType(commenError, commenError, "commenError should have right type")
	expectedError := map[string]interface{}{"database": "no such table: not_exists"}
	assert.Equal(expectedError,
		commenError.Errors, "commenError should have right error info")
}
