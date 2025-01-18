package model

import "gorm.io/gorm"

// RunMigration generates and runs migrations.
func RunMigration(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
