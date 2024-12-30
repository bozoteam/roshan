package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"type:varchar(255);not null"`
	Username     string `gorm:"type:varchar(255);unique;not null"`
	Password     string `gorm:"type:varchar(255);not null"`
	RefreshToken string `gorm:"type:varchar(512);"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
