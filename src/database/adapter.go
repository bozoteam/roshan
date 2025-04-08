package adapter

import (
	"fmt"
	"sync"

	// gorm_postgres "gorm.io/driver/postgres"
	"github.com/bozoteam/roshan/src/helpers"
	gorm_postgres "gorm.io/driver/postgres"

	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

// GetDBConnection returns a singleton instance of a database connection
func GetDBConnection() *gorm.DB {
	var err error

	user := helpers.GetEnv("DB_USER")
	password := helpers.GetEnv("DB_PASSWORD")
	host := helpers.GetEnv("DB_HOST")
	port := helpers.GetEnv("DB_PORT")
	dbName := helpers.GetEnv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbName, port,
	)

	dbInstance, err = gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		err = fmt.Errorf("error opening database connection: %v", err)
		panic(err)
	}

	// return dbInstance
	return dbInstance.Debug()
}
