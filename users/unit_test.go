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
    "regexp"
)

var image_url ="https://golang.org/doc/gopher/frontpage.png"

func newUserModel() UserModel{
    return UserModel{
        ID: 2,
        Username:"asd123!@#ASD",
        Email:"wzt@g.cn",
        Bio:"heheda",
        Image: &image_url,
        JWT:"",
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
    assert.Equal(userModel.JWT, "","JWT's should be null")

    userModel = newUserModel()
    err = userModel.setPassword("")
    assert.Error(err,"empty password can not be set null")
    assert.Equal(userModel.JWT, "","JWT's should be null")

    userModel = newUserModel()
    err = userModel.setPassword("asd123!@#ASD")
    assert.NoError(err,"password should be set successful")
    assert.Len(userModel.PasswordHash, 60,"password hash length should be 60")
    assert.Len(userModel.JWT, 115,"JWT's length should be 115")

    err = userModel.checkPassword("sd123!@#ASD")
    assert.Error(err,"password should be checked and not validated")

    err = userModel.checkPassword("asd123!@#ASD")
    assert.NoError(err,"password should be checked and validated")
    assert.Len(userModel.JWT, 115,"JWT's length should be 115")


    userModel.JWT = "_jwt_jwt"
    marshalModel:=[]byte(`{"id":2,"username":"asd123!@#ASD","email":"wzt@g.cn","bio":"heheda","image":"https://golang.org/doc/gopher/frontpage.png","jwt":"_jwt_jwt"}`)
    modelJSON, _ := json.Marshal(userModel)
    //fmt.Println("%s",string(modelJSON))
    assert.Equal(marshalModel, modelJSON,"Marshal should be equal")

    json.Unmarshal(marshalModel, userModel)
    assertUserModel(t,userModel)
}

func TestRegister(t *testing.T) {
    assert := assert.New(t)

    r := gin.New()
    const path string = "/api/v1/users"

    usersGroup := r.Group(path)
    router := Register(usersGroup)
    assert.Equal(path,router.BasePath,"Base path should be set")
}

func TestRouter_Registration(t *testing.T) {
    assert := assert.New(t)

    test_db, _ := gorm.Open("sqlite3", "./../gorm_test.db")
    test_db.AutoMigrate(&UserModel{})
    defer os.Remove("./../gorm_test.db")

    r := gin.New()
    usersGroup := r.Group("/p")
    usersGroup.Use(middlewares.DatabaseMiddleware(test_db))
    Register(usersGroup)
    var bodyData string= `{"user":{"username": "Jacxxxxxx","email": "wztx@gg.cn","password": "jakejxke"}}`
    //pat := regexp.MustCompile(`\{\"user\":\{\"id\":(\d+),\"username\":\"Jacxxxxxx\",\"email\":\"wztx@gg.cn\",\"bio\":\"\",\"image\":null,\"jwt\":\"(^\"+)\"\}\}`)

    req, err := http.NewRequest("POST", "/p/", bytes.NewBufferString(bodyData))
    req.Header.Set("Content-Type", "application/json")

    assert.NoError(err)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    assert.Equal(http.StatusCreated, w.Code,"response status should be 200")


    pat := regexp.MustCompile(`\{\"user\":\{\"id\":1,\"username\":\"Jacxxxxxx\",\"email\":\"wztx@gg.cn\",\"bio\":\"\",\"image\":null,\"jwt\":\"([^\"]+)\"\}\}`)

    match := pat.FindStringSubmatch(w.Body.String())
    assert.True(len(match)!=0)
    assert.Len(match[1],115,"JWT's length should be 115")

}
