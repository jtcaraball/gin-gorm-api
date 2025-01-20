package model

import (
	"fmt"
	"gin-gorm-api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDBSession returns a DB session as specified by config.
func NewDBSession(config config.Config) (*gorm.DB, error) {
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

// RunMigration generates and runs migrations.
func RunMigration(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
