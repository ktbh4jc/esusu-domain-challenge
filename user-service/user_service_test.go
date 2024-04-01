package user_service

import (
	"encoding/json"
	"errors"
	"maas/loggers"
	"net/http"
	"net/http/httptest"
	"testing"

	error_types "maas/error-types"
	user_model "maas/user-model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
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

	default_user = &user_model.User{
		UserId:          "Danny Default",
		TokensRemaining: 1000,
		IsAdmin:         false,
		AuthKey:         "Danny-Password",
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
	// Note: this couples us to user_model.DefaultUsers
	return allUsers, nil
}
func (m *MockUserRepository) User(id string) (*user_model.User, error) {
	return default_user, nil
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
	m.Run()
}

func testRouter(userService UserService) *gin.Engine {
	router := gin.Default()
	router.GET("/ping", userService.Ping)
	router.POST("/users/reset", userService.ResetDb)
	router.GET("/users/debug", userService.AllUsersDebug)
	router.GET("/users/:id", userService.UserById)
	return router
}

func performRequest(r http.Handler, method string, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

// Unit tests
func TestResetDb_WithNoErrors_ReturnsNewIDsWithStatusOK(t *testing.T) {
	expectedBody := []string{"1", "2", "3"}
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset")

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
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset")
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestResetDb_WithOtherError_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error with DB reset\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "POST", "/users/reset")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestPing_WithNoErrors_ReturnsStatusOK(t *testing.T) {
	expectedBody := "\"Connection good\""

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/ping")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestPing_WithError_RaisesInternalServerError(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error pinging database\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/ping")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAllUsersDebug_WithNoErrors_ReturnsAllUsers(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/debug")

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
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/debug")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUserByID_WithNoErrors_ReturnsUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/660b281d060a0f748cf91f14")

	var response user_model.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, default_user, &response)
}

func TestUserById_WhenUserIsNotFound_RaisesStatusNotFound(t *testing.T) {
	returnedErr := errors.New("test")
	expectedBody := "\"error getting user\""

	mockRepo := &AllErrorsMockUserRepository{}
	mockRepo.setErr(returnedErr)
	service := NewUserService(mockRepo)
	router := testRouter(*service)
	recorder := performRequest(router, "GET", "/users/660b281d060a0f748cf91f14")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}
