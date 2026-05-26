package repository

import (
	"backend/internal/model"
	"backend/pkg/db"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *model.User) error {
	return db.DB.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	if err := db.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(user *model.User) error {
	return db.DB.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return db.DB.Delete(&model.User{}, id).Error
}
