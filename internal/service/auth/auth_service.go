package service

import (
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/repository"
)

const admin_role int = 1
const user_role int = 0

var routes map[string][]accessInfo

type accessInfo struct {
	roles  []int
	method string
}

func init() {
	routes = map[string][]accessInfo{
		"/cocktails": {
			{
				method: "POST",
				roles:  []int{admin_role, user_role},
			},
			{
				method: "GET",
				roles:  []int{admin_role, user_role},
			},
			{
				method: "UPDATE",
				roles:  []int{admin_role, user_role},
			},
			{
				method: "DELETE",
				roles:  []int{admin_role},
			},
		},
	}

}

type AuthService struct {
	authRepository repository.AuthRepository
	userRepository repository.UserRepository
}

func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository) *AuthService {
	return &AuthService{authRepository: authRepo, userRepository: userRepo}
}

func (s *AuthService) AuthUser(token string) (models.User, error) {
	valid, val, err := s.authRepository.ValidateToken(token)
	if err != nil {
		return models.User{}, fmt.Errorf("validatetoken error %s", err.Error())
	}

	if !valid {
		return models.User{}, errors.New("Unauthorized")
	}
	user, err := s.userRepository.GetUser(val)
	if err != nil {
		return models.User{}, fmt.Errorf("getuser error %s", err.Error())
	}
	return user, nil
}

func (s *AuthService) CheckAccess(route string, method string, user models.User) bool {
	value, ok := routes[route]
	if !ok {
		return false
	}
	for _, v := range value {
		if v.method == method {
			for i := 0; i < len(v.roles); i++ {
				if v.roles[i] == user.Role {
					return true
				}
			}
		}
	}
	return false
}
