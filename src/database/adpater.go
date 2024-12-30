package adapter

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

// GetDBConnection returns a singleton instance of a database connection
func GetDBConnection() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		err = godotenv.Load()
		if err != nil {
			err = fmt.Errorf("error loading .env file: %v", err)
			return
		}
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSLMODE")
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbName, port, sslMode)

		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			err = fmt.Errorf("error opening database connection: %v", err)
		}
	})

	return dbInstance, err
}
