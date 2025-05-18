package user_service

import (
	"context"

	gen "github.com/bozoteam/roshan/adapter/grpc/gen/user"
	"github.com/bozoteam/roshan/modules/user/usecase"
)

type UserService struct {
	userUsecase *usecase.UserUsecase
	gen.UnimplementedUserServiceServer
}

func NewUserService(userUsecase *usecase.UserUsecase) *UserService {
	return &UserService{
		userUsecase: userUsecase,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *gen.CreateUserRequest) (*gen.CreateUserResponse, error) {
	user, err := s.userUsecase.CreateUser(ctx, &usecase.UserCreateInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &gen.CreateUserResponse{
		User: &gen.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *gen.UpdateUserRequest) (*gen.UpdateUserResponse, error) {
	user, err := s.userUsecase.UpdateUser(ctx, &usecase.UserUpdateInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &gen.UpdateUserResponse{
		User: &gen.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *gen.DeleteUserRequest) (*gen.DeleteUserResponse, error) {
	user, err := s.userUsecase.DeleteUser(ctx)
	if err != nil {
		return nil, err
	}

	return &gen.DeleteUserResponse{
		User: &gen.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *gen.GetUserRequest) (*gen.GetUserResponse, error) {
	user, err := s.userUsecase.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	return &gen.GetUserResponse{
		User: &gen.User{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}
