package usecase

import (
	"errors"
	"log/slog"

	"context"

	log "github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/helpers"
	"github.com/bozoteam/roshan/modules/user/models"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	"github.com/bozoteam/roshan/roshan_errors"
	"gorm.io/gorm"
)

type UserUsecase struct {
	logger   *slog.Logger
	userRepo *userRepository.UserRepository
}

func NewUserUsecase(db *gorm.DB) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepository.NewUserRepository(db),
		logger:   log.LogWithModule("user_usecase"),
	}
}

// UserCreateInput represents the input for creating a user
type UserCreateInput struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword"`
}

func (c *UserUsecase) CreateUser(ctx context.Context, useReq *UserCreateInput) (*models.User, error) {
	hashedPassword, err := helpers.HashPassword(useReq.Password)
	if err != nil {
		return nil, errors.New("internal server error")
	}

	id := helpers.GenUUID()

	user := &models.User{
		Id:       id,
		Name:     useReq.Name,
		Email:    useReq.Email,
		Password: hashedPassword,
	}

	if err := models.ValidateUser(user); err != nil {
		return nil, roshan_errors.ErrInvalidInput
	}

	if err := c.userRepo.SaveUser(user); err != nil {
		if helpers.IsErrorCode(err, "23505") {
			return nil, errors.New("user email already exists")
		}
		return nil, roshan_errors.ErrInternalServerError
	}

	return user, nil
}

// UserUpdateInput represents the input for updating a user
type UserUpdateInput struct {
	Name     *string `json:"name" example:"John Smith"`
	Email    *string `json:"email" example:"john.smith@example.com"`
	Password *string `json:"password" example:"newsecurepassword"`
}

func (c *UserUsecase) UpdateUser(ctx context.Context, input *UserUpdateInput) (*models.User, error) {
	user := ctx.Value("user").(*models.User)

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		pass, err := helpers.HashPassword(*input.Password)
		if err != nil {
			return nil, roshan_errors.ErrInternalServerError
		}
		user.Password = pass
	}

	if err := models.ValidateUser(user); err != nil {
		return nil, errors.New("Invalid user data")
	}

	if err := c.userRepo.SaveUser(user); err != nil {
		if helpers.IsErrorCode(err, "23505") {
			return nil, errors.New("user already exists")
		}
		return nil, roshan_errors.ErrInternalServerError
	}

	return user, nil
}

func (c *UserUsecase) DeleteUser(ctx context.Context) (*models.User, error) {
	user := ctx.Value("user").(*models.User)

	if err := c.userRepo.DeleteUser(user); err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (c *UserUsecase) GetUser(ctx context.Context) (*models.User, error) {
	user := ctx.Value("user").(*models.User)
	return user, nil
}
