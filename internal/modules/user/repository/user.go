package repository

import (
	"log/slog"

	log "github.com/bozoteam/roshan/internal/adapter/log"
	"github.com/bozoteam/roshan/internal/modules/user/models"
	"gorm.io/gorm"
)

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db, logger: log.LogWithModule("user_repository")}
}

type UserRepository struct {
	logger *slog.Logger
	db     *gorm.DB
}

func (c *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := c.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserRepository) FindUserById(id string) (*models.User, error) {
	var user models.User
	err := c.db.First(&user, "id = ?", id).Error
	if err != nil {
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
