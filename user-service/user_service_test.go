package user_service

import (
	"encoding/json"
	"errors"
	"fmt"
	auth_service "maas/auth-service"
	"maas/loggers"
	"net/http"
	"net/http/httptest"
	"testing"

	error_types "maas/error-types"
	user_model "maas/user-model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	adminIDString   = "111111111111111111111111"
	defaultIDString = "222222222222222222222222"
	otherIDString   = "333333333333333333333333"
)

var (
	authService auth_service.AuthService

	allUsers []user_model.User = []user_model.User{
		{
			UserId:          "Adam Min",
			TokensRemaining: 100,
			IsAdmin:         true,
			AuthKey:         "Super-Secret-Password",
		}, {
			UserId:          "Alice MemeMaster",
			TokensRemaining: 1000,
			IsAdmin:         false,
			AuthKey:         "Alice-MemeMaster-Password",
		}, {
			UserId:          "No-Token Bob",
			TokensRemaining: 0,
			IsAdmin:         false,
			AuthKey:         "Bob-Password",
		},
	}
	adminUser = &user_model.User{
		UserId:          "Adam Min",
		TokensRemaining: 100,
		IsAdmin:         true,
		AuthKey:         "Super-Secret-Password",
	}
	defaultUser = &user_model.User{
		UserId:          "Danny Default",
		TokensRemaining: 1000,
		IsAdmin:         false,
		AuthKey:         "Danny-Password",
	}
	otherUser = &user_model.User{
		UserId:          "Other Ollie",
		TokensRemaining: 1000,
		IsAdmin:         false,
		AuthKey:         "Ollie-Password",
	}
)

// MockUserRepository: Always returns a happy value
type MockUserRepository struct{}

func (m *MockUserRepository) ResetDb() ([]interface{}, error) {
	return []interface{}{"1", "2", "3"}, nil
}
func (m *MockUserRepository) Ping() error {
	return nil
}
func (m *MockUserRepository) AllUsers() ([]user_model.User, error) {
	return allUsers, nil
}
func (m *MockUserRepository) User(id string) (*user_model.User, error) {
	if id == adminIDString {
		return adminUser, nil
	} else if id == defaultIDString {
		return defaultUser, nil
	} else if id == otherIDString {
		return otherUser, nil
	}
	return nil, errors.New("test")
}

func (m *MockUserRepository) UserByAuthHeader(auth string) (*user_model.User, error) {
	if auth == "ADMIN" {
		return adminUser, nil
	} else if auth == "MISSING" {
		return nil, &error_types.AuthUserNotFoundError{}
	} else if auth == "" {
		return nil, &error_types.NoAuthHeaderError{}
	} else {
		return defaultUser, nil
	}
}

// AllErrorsMockUserRepository: Always returns an error
type AllErrorsMockUserRepository struct {
	err error
}

func (m *AllErrorsMockUserRepository) setErr(err error) {
	m.err = err
}

func (m *AllErrorsMockUserRepository) ResetDb() ([]interface{}, error) {
	return []interface{}{}, m.err
}
func (m *AllErrorsMockUserRepository) Ping() error {
	return &error_types.MongoConnectionError{Err: errors.New("test")}
}
func (m *AllErrorsMockUserRepository) AllUsers() ([]user_model.User, error) {
	return []user_model.User{}, errors.New("test")
}
func (m *AllErrorsMockUserRepository) User(id string) (*user_model.User, error) {
	return nil, errors.New("test")
}

// Test utility functions
func TestMain(m *testing.M) {
	loggers.SilentInit()
	setIdHexes()
	authService = *auth_service.NewAuthService(&MockUserRepository{})
	m.Run()
}

func setIdHexes() {
	adminUser = setUserIdHex(adminUser, adminIDString)
	defaultUser = setUserIdHex(defaultUser, defaultIDString)
	otherUser = setUserIdHex(otherUser, otherIDString)
	allUsers[0] = *setUserIdHex(&allUsers[0], adminIDString)
	allUsers[1] = *setUserIdHex(&allUsers[1], defaultIDString)
	allUsers[2] = *setUserIdHex(&allUsers[2], otherIDString)
}

func setUserIdHex(user *user_model.User, id string) *user_model.User {
	hexId, _ := primitive.ObjectIDFromHex(id)
	user.ID = hexId
	return user
}

func testRouter(userService UserService) *gin.Engine {
	router := gin.Default()
	router.GET("/ping", userService.Ping)
	router.POST("/users/reset", userService.ResetDb)
	router.GET("/users/debug", userService.AllUsersDebug)
	router.GET("/users", userService.AllUsers)
	router.GET("/users/:id", userService.UserById)
	return router
}

func performRequest(r http.Handler, method string, path string, authHeader string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("auth", authHeader)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

// Unit tests
func TestResetDb_WithNoErrors_ReturnsNewIDsWithStatusOK(t *testing.T) {
	expectedBody := []string{"1", "2", "3"}
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset", "")

	assert.Equal(t, http.StatusOK, recorder.Code)
	var response []string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, expectedBody, response)
}

func TestResetDb_WithBadEnvError_RaisesBadRequest(t *testing.T) {
	returnedErr := &error_types.BadEnvironmentError{Err: errors.New("test")}
	expectedBody := "\"Unable to reset DB in current environment\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset", "")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestResetDb_WithOtherError_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error with DB reset\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset", "")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestPing_WithNoErrors_ReturnsStatusOK(t *testing.T) {
	expectedBody := "\"Connection good\""

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/ping", "")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestPing_WithError_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error pinging database\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/ping", "")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAllUsers_WhenNoErrors_WhenAdmin_ReturnsUsers(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users", "ADMIN")

	var response []user_model.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, allUsers, response)
}

func TestAllUsers_WhenErrors_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error getting users\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users", "ADMIN")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAllUsers_WhenNotAdmin_RaisesForbidden(t *testing.T) {
	expectedBody := "\"forbidden\""

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users", "DEFAULT")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAllUsers_WhenAuthIsEmpty_RaisesUnauthorized(t *testing.T) {
	expectedBody := "\"unauthorized\""

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users", "")

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAllUsersDebug_WithNoErrors_WithNoAuth_ReturnsAllUsers(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/debug", "")

	var response []user_model.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, allUsers, response)
}

func TestAllUsersDebug_WithErrors_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error getting users\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/debug", "")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUserByID_WhenAdminAsksForAUser_ReturnsUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", defaultIDString), "ADMIN")

	var response user_model.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, defaultUser, &response)
}

func TestUserByID_WhenAUserAsksForThemselves_ReturnsUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", defaultIDString), "DEFAULT")

	var response user_model.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, defaultUser, &response)
}

func TestUserByID_WhenAUserAsksForAnotherUser_RaisesForbidden(t *testing.T) {
	mockRepo := &MockUserRepository{}
	expectedBody := "\"forbidden\""
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", otherIDString), "DEFAULT")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUserByID_WhenAuthHeaderIsNotGiven_RaisesForbidden(t *testing.T) {
	mockRepo := &MockUserRepository{}
	expectedBody := "\"unauthorized\""
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", otherIDString), "")

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUserByID_WhenAuthHeaderIsNotInDB_RaisesUnauthorized(t *testing.T) {
	mockRepo := &MockUserRepository{}
	expectedBody := "\"forbidden\""
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", otherIDString), "MISSING")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUserById_WhenUserIsNotFound_RaisesStatusNotFound(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error getting user\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", fmt.Sprintf("/users/%s", otherIDString), "ADMIN")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}
