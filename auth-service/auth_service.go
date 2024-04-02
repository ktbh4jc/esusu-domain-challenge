package auth_service

import (
	error_types "maas/error-types"
	user_model "maas/user-model"
)

type AuthRepository interface {
	UserByAuthHeader(auth string) (*user_model.User, error)
}

type AuthService struct {
	Repo AuthRepository
}

func NewAuthService(repo AuthRepository) *AuthService {
	return &AuthService{
		Repo: repo,
	}
}

func (s AuthService) IsAdmin(auth string) (bool, error) {
	if auth == "" {
		return false, &error_types.NoAuthHeaderError{}
	}
	user, err := s.Repo.UserByAuthHeader(auth)
	if err != nil {
		return false, err
	}

	return user.IsAdmin, nil
}

// Admins can access any user
func (s AuthService) IsCallerOrAdmin(auth string, id string) (bool, error) {
	if auth == "" {
		return false, &error_types.NoAuthHeaderError{}
	}
	user, err := s.Repo.UserByAuthHeader(auth)
	if err != nil {
		return false, err
	}

	return user.IsAdmin || user.ID.Hex() == id, nil
}

// If a calling user is in the db, we say they are authenticated #securityIsMyPassion
func (s AuthService) IsAuthenticated(auth string) (bool, error) {
	if auth == "" {
		return false, &error_types.NoAuthHeaderError{}
	}
	_, err := s.Repo.UserByAuthHeader(auth)
	if err != nil {
		return false, err
	}
	return true, nil
}
