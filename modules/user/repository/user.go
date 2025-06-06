package repository

import (
	"fmt"
	"log/slog"

	log "github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/modules/user/models"
	"gorm.io/gorm"
)

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db, logger: log.LogWithModule("user_repository")}
}

type UserRepository struct {
	logger *slog.Logger
	db     *gorm.DB
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) SaveRefreshToken(user *models.User, refreshToken string) error {
	if err := r.db.Model(user).Update("refresh_token", refreshToken).Error; err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (r *UserRepository) FindUserById(id string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) SaveUser(user *models.User) error {
	return r.db.Save(user).Error
}

// func (c *UserRepository) UpdateUser(updates map[string]any, user *models.User) error {
// 	return c.db.Model(&user).Updates(updates).Error
// }

// func (c *UserRepository) UpdateUser(updates map[string]any, user *models.User) error {
// 	return c.db.Model(&user).Updates(updates).Error
// }

func (r *UserRepository) FindUserByIdAndToken(id, token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ? AND refresh_token = ?", id, token).First(&user).Error; err != nil {
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

func (r *UserRepository) DeleteUser(user *models.User) error {
	if err := r.db.Delete(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) DeleteRefreshToken(user *models.User) error {
	if err := r.db.Model(user).Update("refresh_token", nil).Error; err != nil {
		return err
	}
	return nil
}
