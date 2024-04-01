package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	meme_maker "maas/meme-maker"
	user_model "maas/user-model"
	user_service "maas/user-service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	userService *user_service.UserService
	router      *gin.Engine
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

func TestMain(m *testing.M) {
	userService = user_service.NewUserService(&MockUserRepository{})
	router = setupRouter(userService)
	m.Run()
}

type MockUserRepository struct{}

func (m *MockUserRepository) ResetDb() ([]interface{}, error)          { panic("unimplemented") }
func (m *MockUserRepository) Ping() error                              { panic("unimplemented") }
func (m *MockUserRepository) AllUsers() ([]user_model.User, error)     { panic("unimplemented") }
func (m *MockUserRepository) User(id string) (*user_model.User, error) { panic("unimplemented") }

// Note: what follows is arguably not the best approach as these are kinda closer to integration tests than unit tests.
// Given time to scale this project up I would focus on splitting my web service tests from my meme maker tests and test the boundaries.
// However since (for now) this project is just a single instance, I am going to do the easy test.
func TestMemeRoute_WithNoParams_RendersDefaultMeme(t *testing.T) {
	expectedBody := meme_maker.NewMeme().MakeMap()

	recorder := performRequest(router, "GET", "/memes")

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.Equal(t, expectedBody, response)
}

func TestMemeRoute_WithGoodParams_RendersCustomMeme(t *testing.T) {
	expectedBody := meme_maker.Meme{TopText: "Test", BottomText: "Bottom Text", ImageLocation: "1.000000 x 2.000000"}

	recorder := performRequest(router, "GET", "/memes?lat=1&lon=2&query=Test")

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, expectedBody.MakeMap(), response)
}

func TestMemeRoute_WithBadParams_RendersBadInput(t *testing.T) {
	expectedBody := map[string]string{"error": "bad request"}

	recorder := performRequest(router, "GET", "/memes?lat=BAD&lon=2&query=Test")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, expectedBody, response)
}
