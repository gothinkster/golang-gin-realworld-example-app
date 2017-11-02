package common

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConnectingDatabase(t *testing.T) {
	assert := assert.New(t)
	db := Init()
	// Test create & close DB
	_, err := os.Stat("./../gorm.db")
	assert.NoError(err, "Db should exist")
	assert.NoError(db.DB().Ping(), "Db should be able to ping")

	// Test get a connecting from connection pools
	connection := GetDB()
	assert.NoError(connection.DB().Ping(), "Db should be able to ping")
	db.Close()


	// Test DB exceptions
	os.Chmod("./../gorm.db",0000)
	db = Init()
	assert.Error(db.DB().Ping(), "Db should not be able to ping")
	db.Close()
	os.Chmod("./../gorm.db",0644)
}


func TestConnectingTestDatabase(t *testing.T) {
	assert := assert.New(t)
	// Test create & close DB
	db := TestDBInit()
	_, err := os.Stat("./../gorm_test.db")
	assert.NoError(err, "Db should exist")
	assert.NoError(db.DB().Ping(), "Db should be able to ping")
	db.Close()

	// Test testDB exceptions
	os.Chmod("./../gorm_test.db",0000)
	db = TestDBInit()
	_, err = os.Stat("./../gorm_test.db")
	assert.NoError(err, "Db should exist")
	assert.Error(db.DB().Ping(), "Db should not be able to ping")
	db.Close()
	os.Chmod("./../gorm_test.db",0644)

	// Test delete DB
	TestDBFree(db)
	_, err = os.Stat("./../gorm_test.db")

	assert.Error(err, "Db should not exist")
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
