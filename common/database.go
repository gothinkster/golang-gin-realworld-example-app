package common

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

var DB *gorm.DB

// GetDBPath returns the database path from environment or default.
// Exported for use in tests.
func GetDBPath() string {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/gorm.db"
	}
	return dbPath
}

// GetTestDBPath returns the test database path from environment or default.
// Exported for use in tests.
func GetTestDBPath() string {
	testDBPath := os.Getenv("TEST_DB_PATH")
	if testDBPath == "" {
		testDBPath = "./data/gorm_test.db"
	}
	return testDBPath
}

// ensureDir creates the directory for the database file if it doesn't exist
func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		return os.MkdirAll(dir, 0750)
	}
	return nil
}

// Opening a database and save the reference to `Database` struct.
func Init() *gorm.DB {
	dbPath := GetDBPath()

	// Ensure the directory exists
	if err := ensureDir(dbPath); err != nil {
		fmt.Println("db err: (Init - create dir) ", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Println("db err: (Init) ", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("db err: (Init - get sql.DB) ", err)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}
	DB = db
	return DB
}

// This function will create a temporarily database for running testing cases
func TestDBInit() *gorm.DB {
	testDBPath := GetTestDBPath()

	// Ensure the directory exists
	if err := ensureDir(testDBPath); err != nil {
		fmt.Println("db err: (TestDBInit - create dir) ", err)
	}

	test_db, err := gorm.Open(sqlite.Open(testDBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Println("db err: (TestDBInit) ", err)
	}
	sqlDB, err := test_db.DB()
	if err != nil {
		fmt.Println("db err: (TestDBInit - get sql.DB) ", err)
	} else {
		sqlDB.SetMaxIdleConns(3)
	}
	DB = test_db
	return DB
}

// Delete the database after running testing cases.
func TestDBFree(test_db *gorm.DB) error {
	sqlDB, err := test_db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}
	testDBPath := GetTestDBPath()
	err = os.Remove(testDBPath)
	return err
}

// Using this function to get a connection, you can create your connection pool here.
func GetDB() *gorm.DB {
	return DB
}
