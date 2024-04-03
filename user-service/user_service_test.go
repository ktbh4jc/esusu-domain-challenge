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
	"maas/models"

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

	allUsers []models.User = []models.User{
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
	otherUser = &models.User{
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
func (m *MockUserRepository) AllUsers() ([]models.User, error) {
	return allUsers, nil
}
func (m *MockUserRepository) User(id string) (*models.User, error) {
	if id == adminIDString {
		return adminUser, nil
	} else if id == defaultIDString {
		return defaultUser, nil
	} else if id == otherIDString {
		return otherUser, nil
	}
	return nil, errors.New("test")
}

func (m *MockUserRepository) UserByAuthHeader(auth string) (*models.User, error) {
	if auth == "ADMIN" {
		return adminUser, nil
	} else if auth == "MISSING" {
		return nil, &error_types.AuthUserNotFoundError{}
	} else if auth == "" {
		return nil, &error_types.NoAuthHeaderError{}
	} else if auth == "AVAILABLE" {
		return nil, &error_types.UnableToLocateDocumentError{}
	} else {
		return defaultUser, nil
	}
}

func (m *MockUserRepository) NewUser(user models.User) (interface{}, error) {
	return "1", nil
}

func (m *MockUserRepository) UpdateUser(id string, user *models.User) error { return nil }

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
func (m *AllErrorsMockUserRepository) AllUsers() ([]models.User, error) {
	return []models.User{}, errors.New("test")
}
func (m *AllErrorsMockUserRepository) User(id string) (*models.User, error) {
	return nil, errors.New("test")
}
func (m *AllErrorsMockUserRepository) NewUser(user models.User) (interface{}, error) {
	panic("Working on it")
}

func (m *AllErrorsMockUserRepository) UpdateUser(id string, user *models.User) error { return m.err }

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

func setUserIdHex(user *models.User, id string) *models.User {
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
	router.POST("/users", userService.NewUser)
	router.GET("/users/:id", userService.UserById)
	router.PATCH("/users/:id", userService.UpdateUser)
	return router
}

func performRequest(r http.Handler, method string, path string, authHeader string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("auth", authHeader)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

func performRequestWithForm(r http.Handler, method string, path string, authHeader string, form map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)

	req.Header.Set("auth", authHeader)
	req.ParseForm()
	for key, value := range form {
		req.PostForm.Set(key, value)
	}
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

	var response []models.User
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

	var response []models.User
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

	var response models.User
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

	var response models.User
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

// There are a few more test cases, but I think this gets us close enough for a takehome.
func TestAddUser_WhenAdminCreatesGoodUser_CreatesUser(t *testing.T) {
	expectedBody := "\"1\""

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "AVAILABLE",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "POST", "/users", "ADMIN", newUser)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAddUser_WhenAdminReusesAuthKey_RaisesError(t *testing.T) {
	expectedBody := fmt.Sprintf("\"User %s is already using that auth header\"", defaultIDString)

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "DEFAULT",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "POST", "/users", "ADMIN", newUser)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestAddUser_WhenNonAdminCreatesUser_RaisesError(t *testing.T) {
	expectedBody := "\"forbidden\""

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "AVAILABLE",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "POST", "/users", "DEFAULT", newUser)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUpdateUser_WhenAdminMakesGoodRequest_ReturnsStatusOK(t *testing.T) {
	expectedBody := "\"successfully updated user\""

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "DEFAULT",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "PATCH", fmt.Sprintf("/users/%s", defaultIDString), "ADMIN", newUser)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUpdateUser_WhenNonAdminMakeRequest_RaisesForbidden(t *testing.T) {
	expectedBody := "\"forbidden\""

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "DEFAULT",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "PATCH", fmt.Sprintf("/users/%s", defaultIDString), "DEFAULT", newUser)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestUpdateUser_WhenUserIdNotInDB_RaisesNotFound(t *testing.T) {
	expectedBody := "\"Unable to find that user\""

	var newUser map[string]string = map[string]string{
		"user_id":          "test_user_id",
		"auth_key":         "DEFAULT",
		"is_admin":         "false",
		"tokens_remaining": "10",
	}

	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo, authService)
	router := testRouter(*service)
	recorder := performRequestWithForm(router, "PATCH", "/users/BAD", "ADMIN", newUser)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, expectedBody, recorder.Body.String())
}
