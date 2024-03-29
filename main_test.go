package main

import (
	"encoding/json"
	meme_maker "maas/meme-maker"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	return recorder
}

// Note: what follows is arguably not the best approach as these are kinda closer to integration tests than unit tests.
// Given time to scale this project up I would focus on splitting my web service tests from my meme maker tests and test the boundaries.
// However since (for now) this project is just a single instance, I am going to do the easy test.
func TestMemeRoute_WithNoParams_RendersDefaultMeme(t *testing.T) {
	expectedBody := meme_maker.NewMeme().MakeMap()

	router := setupRouter()

	recorder := performRequest(router, "GET", "/memes")

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, expectedBody, response)
}

func TestMemeRoute_WithGoodParams_RendersCustomMeme(t *testing.T) {
	expectedBody := meme_maker.Meme{TopText: "Test", BottomText: "Bottom Text", ImageLocation: "1.000000 x 2.000000"}

	router := setupRouter()

	recorder := performRequest(router, "GET", "/memes?lat=1&lon=2&query=Test")

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, expectedBody.MakeMap(), response)
}

func TestMemeRoute_WithBadParams_RendersBadInput(t *testing.T) {
	expectedBody := map[string]string{"error": "bad request"}

	router := setupRouter()

	recorder := performRequest(router, "GET", "/memes?lat=BAD&lon=2&query=Test")

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, expectedBody, response)
}
