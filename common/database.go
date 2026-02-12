package common

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

var DB *gorm.DB

func getDSN() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if host == "" { host = "localhost" }
	if port == "" { port = "5432" }
	if user == "" { user = "postgres" }
	if dbname == "" { dbname = "realworld" }

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)
}

func Init() *gorm.DB {
	dsn := getDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("db err: (Init) ", err)
		panic("Failed to connect to database")
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

func GetDB() *gorm.DB {
	return DB
}

func TestDBInit() *gorm.DB {
	return Init() 
}

func TestDBFree(test_db *gorm.DB) error {
	return nil
}