package users

import (
    "testing"
    "github.com/stretchr/testify/assert"

    "net/http/httptest"
    "encoding/json"
    "gopkg.in/gin-gonic/gin.v1"
    "net/http"
    "bytes"
    "github.com/jinzhu/gorm"
    "os"
    "golang-gin-starter-kit/middlewares"
    _ "regexp"
    _ "fmt"
)

var image_url ="https://golang.org/doc/gopher/frontpage.png"

func newUserModel() UserModel{
    return UserModel{
        ID: 2,
        Username:"asd123!@#ASD",
        Email:"wzt@g.cn",
        Bio:"heheda",
        Image: &image_url,
        Token:"",
        PasswordHash:"",
    }
}
func assertUserModel(t *testing.T, userModel UserModel){
    assert.EqualValues(t, userModel.ID, 2,"Marshal field should be equal")
    assert.Equal(t, *userModel.Image, image_url,"Marshal field should be equal")
    assert.Equal(t, userModel.Username, "asd123!@#ASD","Marshal field should be equal")
    assert.Equal(t, userModel.Email, "wzt@g.cn","Marshal field should be equal")
    assert.Equal(t, userModel.Bio, "heheda","Marshal field should be equal")
}


func TestUsermodel(t *testing.T) {
    assert := assert.New(t)

    userModel := newUserModel()
    err := userModel.checkPassword("")
    assert.Error(err,"empty password should return err")

    userModel = newUserModel()
    err = userModel.setPassword("")
    assert.Error(err,"empty password can not be set null")

    userModel = newUserModel()
    err = userModel.setPassword("asd123!@#ASD")
    assert.NoError(err,"password should be set successful")
    assert.Len(userModel.PasswordHash, 60,"password hash length should be 60")

    err = userModel.checkPassword("sd123!@#ASD")
    assert.Error(err,"password should be checked and not validated")

    err = userModel.checkPassword("asd123!@#ASD")
    assert.NoError(err,"password should be checked and validated")


    userModel.Token = "_token"
    marshalModel:=[]byte(`{"id":2,"username":"asd123!@#ASD","email":"wzt@g.cn","bio":"heheda","image":"https://golang.org/doc/gopher/frontpage.png","token":"_token"}`)
    modelJSON, _ := json.Marshal(userModel)
    //fmt.Println("%s",string(modelJSON))
    assert.Equal(marshalModel, modelJSON,"Marshal should be equal")

    json.Unmarshal(marshalModel, userModel)
    assertUserModel(t,userModel)
}

func TestRegister(t *testing.T) {
    assert := assert.New(t)

    r := gin.New()
    const path string = "/api/users"

    usersGroup := r.Group(path)
    router := UsersRegister(usersGroup)
    assert.Equal(path,router.BasePath,"Base path should be set")
}

var test_db *gorm.DB


var routerRegistrationTests = []struct {
    bodyData        string
    expectedCode    int
    responseRegexg  string
    msg             string
}{
    {
        `{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
        http.StatusCreated,
        `{"user":{"id":1,"username":"wangzitian0","email":"wzt@gg.cn","bio":"","image":null,"token":"([a-zA-Z0-9-_.]{115})"}}`,
        "valid data and should return 200",
    },
    {
        `{"user":{"username": "u","email": "wzt@gg.cn","password": "jakejxke"}}`,
        http.StatusUnprocessableEntity,
        `{"errors":{"Username":"{min: 8}"}}`,
        "short username should return error",
    },
    {
        `{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "j"}}`,
        http.StatusUnprocessableEntity,
        `{"errors":{"Password":"{min: 8}"}}`,
        "short password should return error",
    },
    {
        `{"user":{"username": "wangzitian0","email": "wztgg.cn","password": "jakejxke"}}`,
        http.StatusUnprocessableEntity,
        `{"errors":{"Email":"{key: email}"}}`,
        "email invalid should return error",
    },
}


func TestRouter_Registration(t *testing.T) {
    assert := assert.New(t)


    r := gin.New()
    usersGroup := r.Group("/p")
    usersGroup.Use(middlewares.DatabaseMiddleware(test_db))
    UsersRegister(usersGroup)
    for _, testData := range routerRegistrationTests{
        bodyData := testData.bodyData
        req, err := http.NewRequest("POST", "/p/", bytes.NewBufferString(bodyData))
        req.Header.Set("Content-Type", "application/json")

        assert.NoError(err)
        w := httptest.NewRecorder()
        r.ServeHTTP(w, req)

        assert.Equal(testData.expectedCode, w.Code, "code - " + testData.msg)
        assert.Regexp(testData.responseRegexg, w.Body.String(),"regexp - %v\n " + testData.msg)
    }


}

func TestMain(m *testing.M) {

    test_db, _ = gorm.Open("sqlite3", "./../gorm_test.db")
    test_db.AutoMigrate(&UserModel{})
    //fmt.Println("before")
    exitVal := m.Run()
    //fmt.Println("after")
    test_db.Close()
    os.Remove("./../gorm_test.db")

    os.Exit(exitVal)
}