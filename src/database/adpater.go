package adapter

import (
	"fmt"
	"sync"

	"github.com/bozoteam/roshan/src/helpers"
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
		user := helpers.GetEnv("DB_USER")
		password := helpers.GetEnv("DB_PASSWORD")
		host := helpers.GetEnv("DB_HOST")
		port := helpers.GetEnv("DB_PORT")
		dbName := helpers.GetEnv("DB_NAME")
		sslMode := helpers.GetEnv("DB_SSLMODE")

		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbName, port, sslMode,
		)

		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			err = fmt.Errorf("error opening database connection: %v", err)
		}
	})

	return dbInstance, err
}
