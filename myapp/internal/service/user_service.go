package service

import (
	"context"
	"fmt"

	"myapp/internal/domain"
	"myapp/internal/repository"
)

// UserService はユーザーに関するビジネスロジックを提供するインターフェースです。
type UserService interface {
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	GetUserDetailByID(ctx context.Context, id int) (*domain.UserDetail, error)
}

// userService は UserService の具象実装です。
type userService struct {
	repo repository.UserRepository
}

// NewUserService は UserService の実装を返します。
// 主に Repository を DI(依存注入) するためのコンストラクタとして機能します。
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// GetUserByID は、文字列で受け取ったユーザーIDをパースし、ユーザーを取得して返します。
func (s *userService) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	return user, nil
}

// GetUserDetailByID は、ユーザーIDを受け取り、ユーザーの詳細情報を取得して返します。
func (s *userService) GetUserDetailByID(ctx context.Context, id int) (*domain.UserDetail, error) {
	userDetail, err := s.repo.FindDetailByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user detail by id: %v", err)
	}
	return userDetail, nil
}
