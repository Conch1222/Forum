package service

import (
	"Forum/internal/domain"
	"Forum/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(req *domain.LoginRequest) (string, error) // return JWT token
	GetProfile(userID uint) (*domain.UserResponse, error)
}

type UserServiceImpl struct {
	userRepo repository.UserRepo
	jwtKey   string
}

func NewUserService(repo repository.UserRepo, jwtKey string) UserService {
	return &UserServiceImpl{userRepo: repo, jwtKey: jwtKey}
}

func (s *UserServiceImpl) Register(req *domain.RegisterRequest) (*domain.UserResponse, error) {
	// check if email already exists
	_, err := s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		UserName: req.UserName,
		Email:    req.Email,
		Password: string(hashedPasswd),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return &domain.UserResponse{ID: user.ID, UserName: user.UserName, Email: user.Email}, nil
}

func (s *UserServiceImpl) Login(req *domain.LoginRequest) (string, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(s.jwtKey))
}

func (s *UserServiceImpl) GetProfile(userID uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &domain.UserResponse{ID: user.ID, UserName: user.UserName, Email: user.Email}, nil
}
