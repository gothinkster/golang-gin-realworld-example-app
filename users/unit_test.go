package users

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/gothinkster/golang-gin-realworld-example-app/common"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	_ "regexp"
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
	var offset int
	test_db.Model(&UserModel{}).Count(&offset)
	var ret []UserModel
	for i := offset + 1; i <= offset+n; i++ {
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

//Reset test DB and create new one with mock data
func resetDBWithMock() {
	common.TestDBFree(test_db)
	test_db = common.TestDBInit()
	AutoMigrate()
	userModelMocker(3)
}

func HeaderTokenMock(req *http.Request, u uint) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", common.GenToken(u)))
}

//You could write the init logic like reset database code here
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
		`{"user":{"username":"wangzitian0","email":"wzt@gg.cn","bio":"","image":null,"token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"UNIQUE constraint failed: user_models.email"}}`,
		"duplicated data and should return StatusUnprocessableEntity",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "u","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Username":"{min: 4}"}}`,
		"too short username should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "j"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"too short password should return error",
	},
	{
		func(req *http.Request) {},
		"/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wztgg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Email":"{key: email}"}}`,
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
		http.StatusForbidden,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"email not exist should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "password126"}}`,
		http.StatusForbidden,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"password error should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"password too short should return error info",
	},
	{
		func(req *http.Request) {},
		"/users/login",
		"POST",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
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
			HeaderTokenMock(req, 1)
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
			HeaderTokenMock(req, 1)
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
			HeaderTokenMock(req, 2)
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
			HeaderTokenMock(req, 1)
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
			HeaderTokenMock(req, 1)
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
			HeaderTokenMock(req, 1)
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
			HeaderTokenMock(req, 2)
		},
		"/user/",
		"PUT",
		`{"user":{"password": "pas"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"current user profile should not be changed with error user info",
	},

	//---------------------   Testing for db errors   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			HeaderTokenMock(req, 4)
		},
		"/user/",
		"PUT",
		`{"password": "password321"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Email":"{key: email}","Username":"{key: alphanum}"}}`,
		"test database pk error for user update",
	},
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 0)
		},
		"/user/",
		"PUT",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"UNIQUE constraint failed: user_models.email"}}`,
		"cheat validator and test database connecting error for user update",
	},
	{
		func(req *http.Request) {
			common.TestDBFree(test_db)
			test_db = common.TestDBInit()

			test_db.AutoMigrate(&UserModel{})
			userModelMocker(3)
			HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"no such table: follow_models"}}`,
		"test database error for following",
	},
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"DELETE",
		``,
		http.StatusUnprocessableEntity,
		`{"errors":{"database":"no such table: follow_models"}}`,
		"test database error for canceling following",
	},
	{
		func(req *http.Request) {
			resetDBWithMock()
			HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"POST",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":"Invalid username"}}`,
		"following wrong user name should return errors",
	},
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/profiles/user666/follow",
		"DELETE",
		``,
		http.StatusNotFound,
		`{"errors":{"profile":"Invalid username"}}`,
		"cancel following wrong user name should return errors",
	},

	//---------------------   Testing for user following   ---------------------
	{
		func(req *http.Request) {
			resetDBWithMock()
			HeaderTokenMock(req, 2)
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
			HeaderTokenMock(req, 2)
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
			HeaderTokenMock(req, 2)
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
			HeaderTokenMock(req, 2)
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

//This is a hack way to add test database for each case, as whole test will just share one database.
//You can read TestWithoutAuth's comment to know how to not share database each case.
func TestMain(m *testing.M) {
	test_db = common.TestDBInit()
	AutoMigrate()
	exitVal := m.Run()
	common.TestDBFree(test_db)
	os.Exit(exitVal)
}
