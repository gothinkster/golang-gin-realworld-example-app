package users

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang-gin-starter-kit/common"
	"gopkg.in/gin-gonic/gin.v1"
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
	var ret []UserModel
	for i := 0; i < n; i++ {
		image := fmt.Sprintf("image%v", i)
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

var routerRegistrationTests = []struct {
	url            string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	{
		"/users/",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusCreated,
		`{"user":{"username":"wangzitian0","email":"wzt@gg.cn","bio":"","image":null,"token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},
	{
		"/users/",
		`{"user":{"username": "u","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Username":"{min: 4}"}}`,
		"short username should return error",
	},
	{
		"/users/",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "j"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"short password should return error",
	},
	{
		"/users/",
		`{"user":{"username": "wangzitian0","email": "wztgg.cn","password": "jakejxke"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Email":"{key: email}"}}`,
		"email invalid should return error",
	},
	{
		"/users/login",
		`{"user":{"email": "user1@linkedin.com","password": "password123"}}`,
		http.StatusOK,
		`{"user":{"username":"user1","email":"user1@linkedin.com","bio":"bio1","image":"image1","token":"([a-zA-Z0-9-_.]{115})"}}`,
		"email login should return user",
	},
	{
		"/users/login",
		`{"user":{"email": "user112312312@linkedin.com","password": "password123"}}`,
		http.StatusForbidden,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"email not exist should return error info",
	},
	{
		"/users/login",
		`{"user":{"email": "user1@linkedin.com","password": "password126"}}`,
		http.StatusForbidden,
		`{"errors":{"login":"Not Registered email or invalid password"}}`,
		"password error should return error info",
	},
	{
		"/users/login",
		`{"user":{"email": "user1@linkedin.com","password": "passw"}}`,
		http.StatusUnprocessableEntity,
		`{"errors":{"Password":"{min: 8}"}}`,
		"password too short should return error info",
	},
}

func TestWithoutAuth(t *testing.T) {
	asserts := assert.New(t)

	r := gin.New()
	UsersRegister(r.Group("/users"))
	for _, testData := range routerRegistrationTests {
		bodyData := testData.bodyData
		req, err := http.NewRequest("POST", testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")

		asserts.NoError(err)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
	}
}

//This is a hack way to add test database for each case
func TestMain(m *testing.M) {
	test_db = common.TestDBInit()
	AutoMigrate()
	exitVal := m.Run()
	common.TestDBFree(test_db)
	os.Exit(exitVal)
}
