package auth_service

import (
	error_types "maas/error-types"
	"maas/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	adminIDString   = "111111111111111111111111"
	defaultIDString = "222222222222222222222222"
)

var (
	authService AuthService

	adminUser = &models.User{
		UserId:          "Adam Min",
		TokensRemaining: 100,
		IsAdmin:         true,
		AuthKey:         "Super-Secret-Password",
	}
	defaultUser = &models.User{
		UserId:          "Danny Default",
		TokensRemaining: 1000,
		IsAdmin:         false,
		AuthKey:         "Danny-Password",
	}
)

type MockUserRepository struct{}

func (m *MockUserRepository) UserByAuthHeader(auth string) (*models.User, error) {
	if auth == "ADMIN" {
		return adminUser, nil
	} else if auth == "MISSING" {
		return nil, &error_types.UserNotFoundError{}
	} else if auth == "" {
		return nil, &error_types.NoAuthHeaderError{}
	} else {
		return defaultUser, nil
	}
}

func setUserIdHex(user *models.User, id string) *models.User {
	hexId, _ := primitive.ObjectIDFromHex(id)
	user.ID = hexId
	return user
}

func TestMain(m *testing.M) {
	authService = *NewAuthService(&MockUserRepository{})
	adminUser = setUserIdHex(adminUser, adminIDString)
	defaultUser = setUserIdHex(defaultUser, defaultIDString)
	m.Run()
}

func TestIsAdmin_WhenAdmin_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsAdmin("ADMIN")
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsAdmin_WhenNonAdmin_ReturnsFalseAndNoErrors(t *testing.T) {
	result, err := authService.IsAdmin("DEFAULT")
	assert.Nil(t, err)
	assert.False(t, result)
}

func TestIsAdmin_WhenAuthIsNotInDB_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsAdmin("MISSING")
	assert.ErrorIs(t, err, &error_types.UserNotFoundError{})
	assert.False(t, result)
}

func TestIsAdmin_WhenAuthHeaderIsEmpty_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsAdmin("")
	assert.ErrorIs(t, err, &error_types.NoAuthHeaderError{})
	assert.False(t, result)
}

func TestIsCallerOrAdmin_WhenAdminRequestsThemselves_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("ADMIN", adminIDString)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsCallerOrAdmin_WhenAdminRequestsAnotherUser_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("ADMIN", defaultIDString)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsCallerOrAdmin_WhenNonAdminRequestsThemselves_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("DEFAULT", defaultIDString)
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsCallerOrAdmin_WhenNonAdminRequestsAnotherUser_ReturnsFalseAndNoErrors(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("DEFAULT", adminIDString)
	assert.Nil(t, err)
	assert.False(t, result)
}

func TestIsCallerOrAdmin_WhenAuthIsNotInDB_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("MISSING", adminIDString)
	assert.ErrorIs(t, err, &error_types.UserNotFoundError{})
	assert.False(t, result)
}

func TestIsCallerOrAdmin_WhenAuthHeaderIsEmpty_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsCallerOrAdmin("", adminIDString)
	assert.ErrorIs(t, err, &error_types.NoAuthHeaderError{})
	assert.False(t, result)
}

func TestIsAuthenticated_WhenAdmin_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsAuthenticated("ADMIN")
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsAuthenticated_WhenNonAdmin_ReturnsTrueAndNoErrors(t *testing.T) {
	result, err := authService.IsAuthenticated("DEFAULT")
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestIsAuthenticated_WhenAuthIsNotInDB_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsAuthenticated("MISSING")
	assert.ErrorIs(t, err, &error_types.UserNotFoundError{})
	assert.False(t, result)
}

func TestIsAuthenticated_WhenAuthHeaderIsEmpty_ReturnsFalseUserNotFoundError(t *testing.T) {
	result, err := authService.IsAuthenticated("")
	assert.ErrorIs(t, err, &error_types.NoAuthHeaderError{})
	assert.False(t, result)
}
