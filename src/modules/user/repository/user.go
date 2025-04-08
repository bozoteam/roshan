package repository

import (
	"github.com/bozoteam/roshan/src/modules/user/models"
	"gorm.io/gorm"
)

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

type UserRepository struct {
	db *gorm.DB
}

func (c *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := c.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserRepository) SaveUser(user *models.User) error {
	return c.db.Save(user).Error
}

// func (c *UserRepository) UpdateUser(updates map[string]any, user *models.User) error {
// 	return c.db.Model(&user).Updates(updates).Error
// }

// func (c *UserRepository) UpdateUser(updates map[string]any, user *models.User) error {
// 	return c.db.Model(&user).Updates(updates).Error
// }

func (c *UserRepository) FindUserByIdAndToken(id, token string) (*models.User, error) {
	var user models.User
	if err := c.db.Where("id = ? AND refresh_token = ?", id, token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// func (c *UserRepository) DeleteUserById(id string) error {
// 	if err := c.db.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }

func (c *UserRepository) DeleteUser(user *models.User) error {
	if err := c.db.Delete(user).Error; err != nil {
		return err
	}
	return nil
}
