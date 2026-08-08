package users

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var image_url = "https://golang.org/doc/gopher/frontpage.png"
var test_db *gorm.DB

func newUserModel() UserModel {
	return UserModel{
		ID:           2,
		Username:     "asd123!@#ASD",
		Email:        "wzt@g.cn",
		Bio:          "heheda",
		Image:        &image_url,
		PasswordHash: "",
	}
}

func userModelMocker(n int) []UserModel {
	var offset int64
	test_db.Model(&UserModel{}).Count(&offset)
	var ret []UserModel
	for i := int(offset) + 1; i <= int(offset)+n; i++ {
		image := fmt.Sprintf("http://image/%v.jpg", i)
		userModel := UserModel{
			Username: fmt.Sprintf("user%v", i),
			Email:    fmt.Sprintf("user%v@linkedin.com", i),
			Bio:      fmt.Sprintf("bio%v", i),
			Image:    &image,
		}
		userModel.setPassword("password123")
		test_db.Create(&userModel)
		ret = append(ret, userModel)
	}
	return ret
}

func TestUserModel(t *testing.T) {
	asserts := assert.New(t)

	//Testing UserModel's password feature
	userModel := newUserModel()
	err := userModel.checkPassword("")
	asserts.Error(err, "empty password should return err")

	userModel = newUserModel()
	err = userModel.setPassword("")
	asserts.Error(err, "empty password can not be set null")

	userModel = newUserModel()
	err = userModel.setPassword("asd123!@#ASD")
	asserts.NoError(err, "password should be set successful")
	asserts.Len(userModel.PasswordHash, 60, "password hash length should be 60")

	err = userModel.checkPassword("sd123!@#ASD")
	asserts.Error(err, "password should be checked and not validated")

	err = userModel.checkPassword("asd123!@#ASD")
	asserts.NoError(err, "password should be checked and validated")

	//Testing the following relationship between users
	users := userModelMocker(3)
	a := users[0]
	b := users[1]
	c := users[2]
	asserts.Equal(0, len(a.GetFollowings()), "GetFollowings should be right before following")
	asserts.Equal(false, a.isFollowing(b), "isFollowing relationship should be right at init")
	a.following(b)
	asserts.Equal(1, len(a.GetFollowings()), "GetFollowings should be right after a following b")
	asserts.Equal(true, a.isFollowing(b), "isFollowing should be right after a following b")
	a.following(c)
	asserts.Equal(2, len(a.GetFollowings()), "GetFollowings be right after a following c")
	asserts.EqualValues(b, a.GetFollowings()[0], "GetFollowings should be right")
	asserts.EqualValues(c, a.GetFollowings()[1], "GetFollowings should be right")
	a.unFollowing(b)
	asserts.Equal(1, len(a.GetFollowings()), "GetFollowings should be right after a unFollowing b")
	asserts.EqualValues(c, a.GetFollowings()[0], "GetFollowings should be right after a unFollowing b")
	asserts.Equal(false, a.isFollowing(b), "isFollowing should be right after a unFollowing b")
}

// Reset test DB and create new one with mock data
func resetDBWithMock() {
	common.TestDBFree(test_db)
	test_db = common.TestDBInit()
	AutoMigrate()
	userModelMocker(3)
}

// You could write the init logic like reset database code here
var unauthRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	//Testing will run one by one, so you can combine it to a user story till another init().
	//And you can modified the header or body in the func(req *http.Request) {}

	//---------------------   Testing for user register   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusCreated,
		`{"user":{"username":"wangzitian0","email":"wzt@gg.cn","bio":null,"image":null,"token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusConflict,
		`{"errors":{"username":\["has already been taken"\]}}`,
		"duplicated data and should return StatusConflict",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "u","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"username":\["is too short \(minimum is 4 characters\)"\]}}`,
		"too short username should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "j"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"password":\["is too short \(minimum is 8 characters\)"\]}}`,
		"too short password should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wztgg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"email":\["is invalid"\]}}`,
		"email invalid should return error",
	},

	//---------------------   Testing for user login   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "password123"}}`,
		http.StatusOK,
		`{"user":{"username":"user1","email":"user1@linkedin.com","bio":"bio1","image":"http://image/1.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"right info login should return user",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user112312312@linkedin.com","password": "password123"}}`,
		http.StatusUnauthorized,
		`{"errors":{"credentials":\["invalid"\]}}`,
		"email not exist should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "password126"}}`,
		http.StatusUnauthorized,
		`{"errors":{"credentials":\["invalid"\]}}`,
		"password error should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"password":\["is too short \(minimum is 8 characters\)"\]}}`,
		"password too short should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"password":\["is too short \(minimum is 8 characters\)"\]}}`,
		"password too short should return error info",
	},

	//---------------------   Testing for self info get & auth module  ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/user/",
		"GET",
		``,
		http.StatusUnauthorized,
		``,
		"request should return 401 without token",
	},
	{
		func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("Tokee %v", common.GenToken(1)))
		},
		"/user/",
		"GET",
		``,
		http.StatusUnauthorized,
		``,
		"wrong token should return 401",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/user/",
		"GET",
		``,
		http.StatusOK,
		`{"user":{"username":"user1","email":"user1@linkedin.com","bio":"bio1","image":"http://image/1.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"request should return current user with token",
	},

	//---------------------   Testing for users' profile get   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"anonymous request should return profile with following=false",
	},
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"request should return self profile",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"request should return correct other's profile",
	},

	//---------------------   Testing for users' profile update   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user123",
		"GET",
		``,
		http.StatusNotFound,
		``,
		"user should not exist profile before changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/user/",
		"PUT",
		`{"user":{"username":"user123","password": "password126","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg"}}`,
		http.StatusOK,
		`{"user":{"username":"user123","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"current user profile should be changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 1)
		},
		"/profiles/user123",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user123","bio":"bio123","image":"http://hehe/123.jpg","following":false}}`,
		"request should return self profile after changed",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user123@linkedin.com","password": "password126"}}`,
		http.StatusOK,
		`{"user":{"username":"user123","email":"user123@linkedin.com","bio":"bio123","image":"http://hehe/123.jpg","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"user should login using new password after changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/user/",
		"PUT",
		`{"user":{"password": "pas"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"password":\["is too short \(minimum is 8 characters\)"\]}}`,
		"current user profile should not be changed with error user info",
	},

	//---------------------   Testing for db errors   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 4)
		},
		"/user/",
		"PUT",
		`{"user":{"username": null}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"username":\["can't be blank"\]}}`,
		"null username should be rejected on user update",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 0)
		},
		"/user/",
		"PUT",
		`{"user":`,
		http.StatusUnprocessableEntity,
		`{"errors":{"body":\["is invalid"\]}}`,
		"malformed json body should be rejected on user update",
	},
	{
		func(req *http.Request) {
			common.TestDBFree(test_db)
			test_db = common.TestDBInit()

			test_db.AutoMigrate(&UserModel{})
			userModelMocker(3)
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":\["no such table: follow_models"\]}}`,
		"test database error for following",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"DELETE",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":\["no such table: follow_models"\]}}`,
		"test database error for canceling following",
	},
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"POST",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":\["not found"\]}}`,
		"following wrong user name should return errors",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"DELETE",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":\["not found"\]}}`,
		"cancel following wrong user name should return errors",
	},

	//---------------------   Testing for user following   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":true}}`,
		"user follow another should work",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":true}}`,
		"user follow another should make sure database changed",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"DELETE",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"user cancel follow another should work",
	},
	{
		func(req *http.Request) {
			common.HeaderTokenMock(req, 2)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"bio1","image":"http://image/1.jpg","following":false}}`,
		"user cancel follow another should make sure database changed",
	},
}

func TestWithoutAuth(t *testing.T) {
	asserts := assert.New(t)
	//You could write the reset database code here if you want to create a database for this block
	//resetDB()

	r := gin.New()
	UsersRegister(r.Group("/users"))
	r.Use(AuthMiddleware(false))
	ProfileRetrieveRegister(r.Group("/profiles"))
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	ProfileRegister(r.Group("/profiles"))
	for _, testData := range unauthRequestTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest(testData.method, testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		testData.init(req)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

func TestExtractTokenFromQueryParameter(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(false))
	r.GET("/test", func(c *gin.Context) {
		userID := c.MustGet("my_user_id").(uint)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	resetDBWithMock()

	// Test with access_token query parameter
	token := common.GenToken(1)
	req, _ := http.NewRequest("GET", "/test?access_token="+token, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "Request with query token should succeed")
	asserts.Contains(w.Body.String(), `"user_id":1`, "User ID should be 1")
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Test with invalid JWT token
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token invalid.jwt.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnauthorized, w.Code, "Invalid token should return 401")
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(false))
	r.GET("/test", func(c *gin.Context) {
		userID := c.MustGet("my_user_id").(uint)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Test with no token (auto401=false should still proceed)
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusOK, w.Code, "No token with auto401=false should proceed")
	asserts.Contains(w.Body.String(), `"user_id":0`, "User ID should be 0")
}

// This is a hack way to add test database for each case, as whole test will just share one database.
// You can read TestWithoutAuth's comment to know how to not share database each case.
func TestMain(m *testing.M) {
	test_db = common.TestDBInit()
	AutoMigrate()
	exitVal := m.Run()
	common.TestDBFree(test_db)
	os.Exit(exitVal)
}

// Covers the raw-JSON tri-state semantics of UserUpdate: omitted fields are
// preserved, explicit null clears nullable fields and is rejected for
// required ones, and password rules follow NIST 800-63B.
func TestUserUpdateNullAndBlankSemantics(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	resetDBWithMock()

	doPut := func(body string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("PUT", "/user", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		common.HeaderTokenMock(req, 1)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	w := doPut(`{"user":{"email":null}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "null email should be rejected")
	asserts.Contains(w.Body.String(), `"email":["can't be blank"]`)

	w = doPut(`{"user":{"username":null}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "null username should be rejected")
	asserts.Contains(w.Body.String(), `"username":["can't be blank"]`)

	w = doPut(`{"user":{"email":"not-an-email"}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "malformed email should be rejected")
	asserts.Contains(w.Body.String(), `"email":["is invalid"]`)

	w = doPut(`{"user":{"password":""}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "blank password should be rejected")
	asserts.Contains(w.Body.String(), `"password":["can't be blank"]`)

	w = doPut(`{"user":{"password":"short7c"}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "7-char password should be rejected")
	asserts.Contains(w.Body.String(), "is too short")

	longPassword := string(bytes.Repeat([]byte("a"), 256))
	w = doPut(fmt.Sprintf(`{"user":{"password":"%s"}}`, longPassword))
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "256-char password should be rejected")
	asserts.Contains(w.Body.String(), "is too long")

	w = doPut(`{"user":{"bio":123}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "non-string bio should be rejected")
	asserts.Contains(w.Body.String(), `"bio":["is invalid"]`)

	w = doPut(`{"user":{"image":[1]}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "non-string image should be rejected")
	asserts.Contains(w.Body.String(), `"image":["is invalid"]`)

	w = doPut(`{"user":{"username":123}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "non-string username should be rejected")
	asserts.Contains(w.Body.String(), `"username":["is invalid"]`)

	// A wrong-typed field is reported as invalid alongside other field errors
	w = doPut(`{"user":{"email":null,"bio":123}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "null email with invalid bio should be rejected")
	asserts.Contains(w.Body.String(), `"email":["can't be blank"]`)
	asserts.Contains(w.Body.String(), `"bio":["is invalid"]`)

	w = doPut(`{"user":`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "malformed JSON should be rejected")
	asserts.Contains(w.Body.String(), `"body":["is invalid"]`)

	// Nullable fields: explicit null and empty string both clear to null
	w = doPut(`{"user":{"bio":null,"image":""}}`)
	asserts.Equal(http.StatusOK, w.Code, "clearing bio and image should succeed")
	asserts.Contains(w.Body.String(), `"bio":null`)
	asserts.Contains(w.Body.String(), `"image":null`)

	// Omitted fields are preserved
	w = doPut(`{"user":{"bio":"kept bio"}}`)
	asserts.Equal(http.StatusOK, w.Code, "partial update should succeed")
	asserts.Contains(w.Body.String(), `"username":"user1"`)
	asserts.Contains(w.Body.String(), `"bio":"kept bio"`)

	// Valid password change is accepted (64 chars per NIST must be accepted)
	okPassword := string(bytes.Repeat([]byte("a"), 64))
	w = doPut(fmt.Sprintf(`{"user":{"password":"%s"}}`, okPassword))
	asserts.Equal(http.StatusOK, w.Code, "64-char password should be accepted")
}

// Covers the 409 duplicate-email branch of UsersRegistration (the duplicate
// username branch is exercised by the table-driven tests above).
func TestUsersRegistrationDuplicateEmail(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	UsersRegister(r.Group("/users"))
	resetDBWithMock()

	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(
		`{"user":{"username":"freshuser","email":"user1@linkedin.com","password":"password123"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusConflict, w.Code, "duplicate email should return 409")
	asserts.Contains(w.Body.String(), `"email":["has already been taken"]`)
}

func TestUsersRegistrationWithImage(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	UsersRegister(r.Group("/users"))
	resetDBWithMock()

	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(
		`{"user":{"username":"imageuser","email":"imageuser@example.com","password":"password123","image":"http://image/profile.jpg"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusCreated, w.Code, "registration with an image should succeed")
	userModel, err := FindOneUser(&UserModel{Username: "imageuser"})
	asserts.NoError(err)
	asserts.NotNil(userModel.Image)
	asserts.Equal("http://image/profile.jpg", *userModel.Image)
}

func TestUsersRegistrationOverlongPasswordRejected(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	UsersRegister(r.Group("/users"))
	resetDBWithMock()

	// bcrypt rejects passwords longer than 72 bytes; the error must surface
	// as a 422 instead of silently storing an unusable hash.
	password := string(bytes.Repeat([]byte("a"), 100))
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(fmt.Sprintf(
		`{"user":{"username":"longpwuser","email":"longpw@example.com","password":"%s"}}`, password)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "over-72-byte password should be rejected")
	asserts.Contains(w.Body.String(), `"body":["is invalid"]`)
	_, err := FindOneUser(&UserModel{Username: "longpwuser"})
	asserts.Error(err, "no user should be created")
}

func TestUserUpdateWrongTypedIdentityFields(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	resetDBWithMock()

	doPut := func(body string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("PUT", "/user", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		common.HeaderTokenMock(req, 1)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	w := doPut(`{"user":{"email":123}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "non-string email should be rejected")
	asserts.Contains(w.Body.String(), `"email":["is invalid"]`)

	w = doPut(`{"user":{"password":123}}`)
	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "non-string password should be rejected")
	asserts.Contains(w.Body.String(), `"password":["is invalid"]`)
}

func TestUserUpdateOverlongPasswordRejected(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	resetDBWithMock()

	// Long enough for the binding tags (max=255) but over bcrypt's 72-byte cap
	password := string(bytes.Repeat([]byte("a"), 100))
	req, _ := http.NewRequest("PUT", "/user", bytes.NewBufferString(fmt.Sprintf(
		`{"user":{"password":"%s"}}`, password)))
	req.Header.Set("Content-Type", "application/json")
	common.HeaderTokenMock(req, 1)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.Equal(http.StatusUnprocessableEntity, w.Code, "over-72-byte password should be rejected")
	asserts.Contains(w.Body.String(), `"password"`)
}

func TestAuthMiddlewareRejectsNonHMACToken(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	r.Use(AuthMiddleware(true))
	UserRegister(r.Group("/user"))
	resetDBWithMock()

	// An unsigned ("none" algorithm) token must not pass the HMAC check
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"id": 1})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	asserts.NoError(err)

	req, _ := http.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", "Token "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	asserts.Equal(http.StatusUnauthorized, w.Code, "non-HMAC token should be rejected")
}

// Simulates the registration race deterministically: a one-shot gorm callback
// plays the concurrent request by inserting a conflicting email right before
// the handler's own INSERT runs, after the pre-checks have already passed.
func TestUsersRegistrationEmailRaceConflict(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	UsersRegister(r.Group("/users"))
	resetDBWithMock()

	db := common.GetDB()
	raced := false
	err := db.Callback().Create().Before("gorm:create").Register("test:email_race", func(tx *gorm.DB) {
		if raced {
			return
		}
		if _, ok := tx.Statement.Dest.(*UserModel); !ok {
			return
		}
		raced = true
		tx.Session(&gorm.Session{NewDB: true}).Exec("INSERT INTO user_models (username, email, bio, password) VALUES (?, ?, ?, ?)",
			"racewinner", "raced@example.com", "", "hash")
	})
	asserts.NoError(err)
	defer db.Callback().Create().Remove("test:email_race")

	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(
		`{"user":{"username":"raceloser","email":"raced@example.com","password":"password123"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	asserts.True(raced, "the simulated concurrent insert should have run")
	asserts.Equal(http.StatusConflict, w.Code, "losing the insert race should return 409")
	asserts.Contains(w.Body.String(), `"email":["has already been taken"]`)
}
