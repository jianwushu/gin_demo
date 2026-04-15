package service

import (
	"context"
	"errors"
	"fmt"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	user := &model.User{
		Name:  req.Name,
		Email: req.Email,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, formatDBError(err)
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, formatDBError(err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, formatDBError(err)
	}

	user.Name = req.Name
	user.Email = req.Email
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, formatDBError(err)
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return formatDBError(err)
	}
	return s.repo.Delete(ctx, id)
}

func formatDBError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("user not found")
	}
	return err
}
