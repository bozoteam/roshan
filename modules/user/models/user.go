package models

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Id       string `validate:"uuid" json:"id" gorm:"primaryKey"`
	Email    string `validate:"email,max=255" gorm:"type:varchar(255);unique;not null" json:"email"`
	Name     string `validate:"required,alphanumunicode,max=255" gorm:"type:varchar(255);not null" json:"name"`
	Password string `validate:"required,ascii,max=72" json:"-" gorm:"type:varchar(72);not null"`

	RefreshToken string    `json:"-" gorm:"type:varchar(1024);"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (u *User) Clone() *User {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(u)
	if err != nil {
		panic(err)
	}

	var result User
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	return &result
}

var modelValidator = validator.New()

func (User) TableName() string {
	return "user"
}

func ValidateUser(user *User) error {
	return modelValidator.Struct(user)
}
