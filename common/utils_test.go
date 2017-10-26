package common

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	assert := assert.New(t)
	db := Init()
	defer db.Close()

	test_db, _ := gorm.Open("sqlite3", "./../gorm_test.db")
	defer test_db.Close()
	assert.IsType(test_db, db, "Db'type should be gorm.DB")
	assert.NoError(db.DB().Ping(), "Db should be able to ping")
	assert.NoError(test_db.DB().Ping(), "Test Db should be able to ping")

	var err = os.Remove("./../gorm_test.db")
	assert.NoError(err, "Db should be deleted")
}

func TestRandString(t *testing.T) {
	assert := assert.New(t)

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	str := RandString(0)
	assert.Equal(str, "", "length should be ''")

	str = RandString(10)
	assert.Equal(len(str), 10, "length should be 10")
	for _, ch := range str {
		assert.Contains(letters, ch, "char should be a-z|A-Z|0-9")
	}
}

func TestGenToken(t *testing.T) {
	assert := assert.New(t)

	token := GenToken(2)

	assert.IsType(token, string("token"), "token type should be string")
	assert.Len(token, 115, "JWT's length should be 115")
}
