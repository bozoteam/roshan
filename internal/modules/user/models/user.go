package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Id       string `validate:"uuid" json:"-" gorm:"primaryKey"`
	Email    string `validate:"email,max=255" gorm:"type:varchar(255);unique;not null" json:"email"`
	Name     string `validate:"required,alphanumunicode,max=255" gorm:"type:varchar(255);not null" json:"name"`
	Password string `validate:"required,ascii,max=60" json:"-" gorm:"type:varchar(60);not null"`

	RefreshToken string    `json:"-" gorm:"type:varchar(1024);"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

var modelValidator = validator.New()

func (User) TableName() string {
	return "user"
}

func ValidateUser(user *User) error {
	return modelValidator.Struct(user)
}
