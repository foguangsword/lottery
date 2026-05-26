package service

import (
	"backend/internal/model"
	"backend/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(username, email, address string) (*model.User, error) {
	user := &model.User{
		Username: username,
		Email:    email,
		Address:  address,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUser(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) ListUsers() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}
