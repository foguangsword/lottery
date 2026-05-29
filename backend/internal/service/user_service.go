package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(username, email, password, ethaddress string) (*model.User, error) {
	hashPwd, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:   username,
		Email:      email,
		Password:   hashPwd,
		EthAddress: ethaddress,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(username, password, jwtSecret string) (string, error) {
	user, err := s.repo.FindByName(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户不存在")
		}
		return "", errors.New("系统错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("密码错误")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func (s *UserService) GetUser(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.repo.FindByEmail(email)
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

func hashPassword(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(bytes), err
}
