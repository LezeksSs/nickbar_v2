package service

import (
	"errors"
	"nickbar_v2/internal/mocks"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
	"nickbar_v2/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newUserSvc(t *testing.T) (*UserService, *mocks.UserRepository, *mocks.AuthRepository) {
	repoUser := mocks.NewUserRepository(t)
	repoAuth := mocks.NewAuthRepository(t)
	return NewUserService(repoUser, repoAuth), repoUser, repoAuth
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		setup    func(u *mocks.UserRepository, a *mocks.AuthRepository, nick string)
		wantTok  repository.TokenPayload
		wantErr  bool
	}{
		{
			name:     "success",
			nickname: "john",
			setup: func(u *mocks.UserRepository, a *mocks.AuthRepository, nick string) {
				u.EXPECT().
					CreateUser(nick).
					Return(models.User{Nickname: nick}, nil)

				a.EXPECT().
					ProduceToken(nick).
					Return(repository.TokenPayload{
						AccessToken:  "acc",
						RefreshToken: "ref",
					}, nil)
			},
			wantTok: repository.TokenPayload{
				AccessToken:  "acc",
				RefreshToken: "ref",
			},
		},
		{
			name:     "create user error",
			nickname: "error_user",
			setup: func(u *mocks.UserRepository, a *mocks.AuthRepository, nick string) {
				u.EXPECT().
					CreateUser(nick).
					Return(models.User{}, errors.New("db fail"))
				// ProduceToken НЕ должен вызываться
			},
			wantErr: true,
		},
		{
			name:     "produce token error",
			nickname: "alice",
			setup: func(u *mocks.UserRepository, a *mocks.AuthRepository, nick string) {
				u.EXPECT().
					CreateUser(nick).
					Return(models.User{Nickname: nick}, nil)

				a.EXPECT().
					ProduceToken(nick).
					Return(repository.TokenPayload{}, errors.New("jwt fail"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, userRepo, authRepo := newUserSvc(t)

			if tc.setup != nil {
				tc.setup(userRepo, authRepo, tc.nickname)
			}

			got, err := svc.RegisterUser(requests.LoginRequest{Nickname: tc.nickname})

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTok, got)
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		setup    func(a *mocks.AuthRepository, nick string)
		wantTok  repository.TokenPayload
		wantErr  bool
	}{
		{
			name:     "success login",
			nickname: "john",
			setup: func(a *mocks.AuthRepository, nick string) {
				a.EXPECT().
					ProduceToken(nick).
					Return(repository.TokenPayload{
						AccessToken:  "acc-token",
						RefreshToken: "ref-token",
					}, nil)
			},
			wantTok: repository.TokenPayload{
				AccessToken:  "acc-token",
				RefreshToken: "ref-token",
			},
		},
		{
			name:     "token generation fails",
			nickname: "invalid",
			setup: func(a *mocks.AuthRepository, nick string) {
				a.EXPECT().
					ProduceToken(nick).
					Return(repository.TokenPayload{}, errors.New("token error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, _, authRepo := newUserSvc(t)

			tc.setup(authRepo, tc.nickname)

			got, err := svc.LoginUser(requests.LoginRequest{Nickname: tc.nickname})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTok, got)
			}
		})
	}
}

func TestGetUserInformation(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		setup    func(repo *mocks.UserRepository, nickname string)
		wantUser models.User
		wantErr  bool
	}{
		{
			name:     "user exists",
			nickname: "john",
			setup: func(repo *mocks.UserRepository, nickname string) {
				repo.EXPECT().
					GetUser(nickname).
					Return(models.User{
						Id:       "u1",
						Nickname: "john",
						Role:     1,
					}, nil)
			},
			wantUser: models.User{
				Id:       "u1",
				Nickname: "john",
				Role:     1,
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			nickname: "ghost",
			setup: func(repo *mocks.UserRepository, nickname string) {
				repo.EXPECT().
					GetUser(nickname).
					Return(models.User{}, errors.New("not found"))
			},
			wantUser: models.User{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			svc, userRepo, _ := newUserSvc(t)

			tt.setup(userRepo, tt.nickname)

			got, err := svc.GetUserInformation(tt.nickname)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, got)
			}
		})
	}
}

func TestChangeUsersPicture(t *testing.T) {
	tests := []struct {
		name      string
		user      models.User
		setup     func(repo *mocks.UserRepository, user models.User)
		expectErr bool
	}{
		{
			name: "successful update",
			user: models.User{
				Id:      "u1",
				Picture: "new_pic.png",
			},
			setup: func(repo *mocks.UserRepository, user models.User) {
				repo.EXPECT().
					UpdatePicture(user).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "update fails",
			user: models.User{
				Id:      "u2",
				Picture: "bad_pic.png",
			},
			setup: func(repo *mocks.UserRepository, user models.User) {
				repo.EXPECT().
					UpdatePicture(user).
					Return(errors.New("db error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			svc, userRepo, _ := newUserSvc(t)

			tt.setup(userRepo, tt.user)

			err := svc.ChangeUsersPicture(tt.user)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChangeUsersRole(t *testing.T) {
	tests := []struct {
		name      string
		user      models.User
		setup     func(repo *mocks.UserRepository, user models.User)
		expectErr bool
	}{
		{
			name: "success",
			user: models.User{Id: "u1", Role: 1},
			setup: func(repo *mocks.UserRepository, user models.User) {
				repo.EXPECT().UpdateRole(user).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "repository error",
			user: models.User{Id: "u2", Role: 2},
			setup: func(repo *mocks.UserRepository, user models.User) {
				repo.EXPECT().UpdateRole(user).Return(errors.New("db error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, userRepo, _ := newUserSvc(t)

			tt.setup(userRepo, tt.user)

			err := svc.ChangeUsersRole(tt.user)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		nickname  string
		setup     func(repo *mocks.UserRepository, nickname string)
		expectErr bool
	}{
		{
			name:     "success",
			nickname: "user1",
			setup: func(repo *mocks.UserRepository, nickname string) {
				repo.EXPECT().DeleteUser(nickname).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "repository error",
			nickname: "user2",
			setup: func(repo *mocks.UserRepository, nickname string) {
				repo.EXPECT().DeleteUser(nickname).Return(errors.New("delete failed"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, userRepo, _ := newUserSvc(t)

			tt.setup(userRepo, tt.nickname)

			err := svc.DeleteUser(tt.nickname)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShowAllUsers(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(repo *mocks.UserRepository)
		expected  []models.User
		expectErr bool
	}{
		{
			name: "success",
			setup: func(repo *mocks.UserRepository) {
				repo.EXPECT().GetUserList().Return([]models.User{
					{Id: "1", Nickname: "User1"},
					{Id: "2", Nickname: "User2"},
				}, nil)
			},
			expected: []models.User{
				{Id: "1", Nickname: "User1"},
				{Id: "2", Nickname: "User2"},
			},
			expectErr: false,
		},
		{
			name: "repository error",
			setup: func(repo *mocks.UserRepository) {
				repo.EXPECT().GetUserList().Return(nil, errors.New("db error"))
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, userRepo, _ := newUserSvc(t)

			tt.setup(userRepo)

			users, err := svc.ShowAllUsers()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *users)
			}
		})
	}
}
