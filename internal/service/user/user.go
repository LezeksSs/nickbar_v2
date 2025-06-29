package service

import (
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
	"nickbar_v2/internal/repository"
)

type UserService struct {
	userRepository UserRepository
	authRepository AuthRepository
}

type UserRepository interface {
	GetUser(nickname string) (models.User, error)
	CreateUser(nickname string) (models.User, error)
	UpdatePicture(user models.User) error
	UpdateRole(user models.User) error
	DeleteUser(nickname string) error
	GetUserList() ([]models.User, error)
}

type AuthRepository interface {
	ProduceToken(nickname string) (repository.TokenPayload, error)
	ValidateToken(accToken string) (bool, string, error)
	RefreshTokens(refrToken string) error
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) RegisterNewUser(nickname requests.LoginRequest) (repository.TokenPayload, error) {
	_, err := s.userRepository.CreateUser(nickname.Nickname)

	if err != nil {
		return repository.TokenPayload{}, err
	}

	tokenPayload, err1 := s.authRepository.ProduceToken(nickname.Nickname)
	if err1 != nil {
		return repository.TokenPayload{}, err1
	}

	return tokenPayload, nil
}

func (s *UserService) LoginUser(nickname requests.LoginRequest) (repository.TokenPayload, error) {
	tokenPayload, err1 := s.authRepository.ProduceToken(nickname.Nickname)
	if err1 != nil {
		return repository.TokenPayload{}, err1
	}

	return tokenPayload, nil
}

func (s *UserService) GetUserInformation(nickname string) (models.User, error) {
	user, err := s.userRepository.GetUser(nickname)

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s *UserService) ChangeUsersPicture(u models.User) error {
	err := s.userRepository.UpdatePicture(u)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ChangeUsersRole(u models.User) error {
	err := s.userRepository.UpdateRole(u)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(nickname string) error {
	err := s.userRepository.DeleteUser(nickname)

	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ShowAllUsers() (*[]models.User, error) {
	users, err := s.userRepository.GetUserList()

	if err != nil {
		return nil, err
	}

	return &users, nil
}
