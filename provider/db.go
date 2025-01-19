package provider

import (
	"fmt"
	"gin-gorm-api/server"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB returns a DB session as specified by config.
func ConnectDB(config server.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.DB.Host,
		config.DB.User,
		config.DB.Password,
		config.DB.Name,
		config.DB.Port,
		config.DB.SSL,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
