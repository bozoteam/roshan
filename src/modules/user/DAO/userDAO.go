package userDAO

import (
	adapter "github.com/bozoteam/roshan/src/database"
	"github.com/bozoteam/roshan/src/helpers"
	"github.com/bozoteam/roshan/src/modules/user/models"
)

func FindUserByUsername(username string) (*models.User, error) {
	db := adapter.GetDBConnection()
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func SaveUser(user *models.User) error {
	db := adapter.GetDBConnection()
	return db.Save(user).Error
}

func FindUserByUsernameAndToken(username, token string) (*models.User, error) {
	db := adapter.GetDBConnection()
	var user models.User
	if err := db.Where("username = ? AND refresh_token = ?", username, token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	db := adapter.GetDBConnection()
	return db.Create(user).Error
}

func UpdateUser(username string, updates map[string]interface{}) error {
	db := adapter.GetDBConnection()
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return err
	}
	if _, ok := updates["password"]; ok {
		hashedPassword, err := helpers.HashPassword(updates["password"].(string))
		if err != nil {
			return err
		}
		updates["password"] = hashedPassword
	}
	return db.Model(&user).Updates(updates).Error
}

func DeleteUserByUsername(username string) error {
	db := adapter.GetDBConnection()
	if err := db.Where("username = ?", username).Delete(&models.User{}).Error; err != nil {
		return err
	}
	return nil
}
