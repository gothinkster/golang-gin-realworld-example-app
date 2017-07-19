package users

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "encoding/json"
)

func TestUsermodel(t *testing.T) {
    assert := assert.New(t)
    image := "https://golang.org/doc/gopher/frontpage.png"
    userModel := UserModel{
        ID: 2,
        Username:"asd123!@#ASD",
        Email:"wzt@g.cn",
        Bio:"heheda",
        Image: &image,
    }
    userModel.JWT = ""
    err := userModel.checkPassword("")
    assert.Error(err,"empty password should return err")
    assert.Equal(userModel.JWT, "","JWT's should be null")

    userModel.JWT = ""
    err = userModel.setPassword("")
    assert.Error(err,"empty password can not be set null")
    assert.Equal(userModel.JWT, "","JWT's should be null")

    userModel.JWT = ""
    err = userModel.setPassword("asd123!@#ASD")
    assert.NoError(err,"password should be set successful")
    assert.Len(userModel.PasswordHash, 60,"password hash length should be 60")
    assert.Len(userModel.JWT, 115,"JWT's length should be 115")

    userModel.JWT = ""
    err = userModel.checkPassword("asd123!@#ASD")
    assert.NoError(err,"password should be checked and validated")
    assert.Len(userModel.JWT, 115,"JWT's length should be 115")

    userModel.JWT = ""
    err = userModel.checkPassword("sd123!@#ASD")
    assert.Error(err,"password should be checked and not validated")
    assert.Equal("", userModel.JWT,"JWT's should be null")


    userModel.JWT = "_jwt_jwt"
    modelJSON, _ := json.Marshal(userModel)
    assert.Equal("{\"id\":2,\"username\":\"asd123!@#ASD\",\"email\":\"wzt@g.cn\",\"bio\":\"heheda\",\"image\":\"https://golang.org/doc/gopher/frontpage.png\",\"jwt\":\"_jwt_jwt\"}",
        string(modelJSON),"Marshal should be equal")
}
