package meme_service

import (
	"encoding/json"
	"errors"
	auth_service "maas/auth-service"
	error_types "maas/error-types"
	"maas/loggers"
	"maas/models"
	"net/http"
	"net/http/httptest"
	"testing"

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
	memeService MemeService

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
		TokensRemaining: 0,
		IsAdmin:         false,
		AuthKey:         "Ollie-Password",
	}

	defaultMeme = &models.Meme{
		TopText:       "Up Top",
		BottomText:    "Down Low",
		ImageLocation: "Nowhere and everywhere",
	}

	paramMeme = &models.Meme{
		TopText:       "Detected input",
		BottomText:    "Bottom Text",
		ImageLocation: "In a full implementation I would like this to be a  URI for an image in an s3 bucket or similar",
	}
)

type MockUserRepository struct{}

func (m *MockUserRepository) UpdateUser(id string, user *models.User) error {
	return nil
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
	} else if auth == "THROW" {
		return nil, errors.New("TEST")
	} else if auth == "OTHER" {
		return otherUser, nil
	} else {
		return defaultUser, nil
	}
}

type MockMemeProvider struct{}

// BuildMeme implements MemeProvider.
func (m *MockMemeProvider) BuildMeme(params *QueryParams) (*models.Meme, error) {
	if params.Query == "someQuery" {
		return paramMeme, nil
	} else if params.Query == "raiseError" {
		return nil, errors.New("test")
	} else {
		return defaultMeme, nil
	}
}

func TestMain(m *testing.M) {
	loggers.SilentInit()
	setIdHexes()
	authService = *auth_service.NewAuthService(&MockUserRepository{})
	memeService = *NewMemeService(&MockUserRepository{}, authService, &MockMemeProvider{})
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

func buildTestContext(path string) *gin.Context {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request, _ = http.NewRequest("GET", path, nil)
	return context
}

func testRouter(memeService MemeService) *gin.Engine {
	router := gin.Default()
	router.GET("/meme", memeService.GetMeme)
	return router
}

func performRequest(r http.Handler, method string, path string, authHeader string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("auth", authHeader)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

func TestGetMeme_WhenEverythingIsGood_RendersAMeme(t *testing.T) {
	expected_body := *defaultMeme
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme", "ADMIN")

	assert.Equal(t, http.StatusOK, recorder.Code)
	var response models.Meme
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, expected_body, response)
}

func TestGetMeme_WhenParamsAreProvided_RendersACustomMeme(t *testing.T) {
	expected_body := *paramMeme
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme?query=someQuery", "ADMIN")

	assert.Equal(t, http.StatusOK, recorder.Code)
	var response models.Meme
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, expected_body, response)
}

func TestGetMeme_WhenBadParamsAreProvided_RaisesAnError(t *testing.T) {
	expected_body := "bad request"
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme?query=someQuery&lat=BAD", "ADMIN")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), expected_body)
}

func TestGetMeme_WhenMemeProviderErrors_RaisesAnError(t *testing.T) {
	expected_body := "Unable to make meme"
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme?query=raiseError", "ADMIN")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	assert.Contains(t, recorder.Body.String(), expected_body)
}

func TestGetMeme_WhenAuthHeaderIsNotProvided_RaisesAnError(t *testing.T) {
	expected_body := "\"unauthorized\""
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme", "")

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, expected_body, recorder.Body.String())
}

func TestGetMeme_WhenAuthHeaderDoesNotMatchAUser_RaisesAnError(t *testing.T) {
	expected_body := "\"forbidden\""
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme", "MISSING")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expected_body, recorder.Body.String())
}

func TestGetMeme_WhenAuthThrowsUnknownError_RaisesAnError(t *testing.T) {
	expected_body := "\"forbidden\""
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme", "THROW")

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Equal(t, expected_body, recorder.Body.String())
}

func TestGetMeme_WhenUserIsOutOfTokens_RaisesAnError(t *testing.T) {
	expected_body := "Tokens needed to make more memes. Buy some!"
	router := testRouter(memeService)
	recorder := performRequest(router, "GET", "/meme", "OTHER")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), expected_body)
}

func TestExtractParams_WithNoParams_ReturnsZeroValueParams(t *testing.T) {
	path := "/test"

	context := buildTestContext(path)
	params, err := memeService.ExtractParams(context)
	assert.Nil(t, err)
	assert.Equal(t, &QueryParams{}, params)
}

func TestExtractParams_WithParams_SetsParams(t *testing.T) {
	path := "/test?query=test&lat=1&lon=2"

	context := buildTestContext(path)
	params, err := memeService.ExtractParams(context)
	assert.Nil(t, err)
	assert.Equal(t, &QueryParams{Query: "test", Lat: 1, Lon: 2}, params)
}

func TestExtractParams_WithPartialParams_SetsParams(t *testing.T) {
	path := "/test?query=test"

	context := buildTestContext(path)
	params, err := memeService.ExtractParams(context)
	assert.Nil(t, err)
	assert.Equal(t, &QueryParams{Query: "test"}, params)
}

func TestExtractParams_WithBadLon_ThrowsError(t *testing.T) {
	path := "/test?lon=five"

	context := buildTestContext(path)
	params, err := memeService.ExtractParams(context)
	assert.Error(t, err)
	assert.Nil(t, params)
}

func TestExtractParams_WithBadLat_ThrowsError(t *testing.T) {
	path := "/test?lat=lon"

	context := buildTestContext(path)
	params, err := memeService.ExtractParams(context)
	assert.Error(t, err)
	assert.Nil(t, params)
}
