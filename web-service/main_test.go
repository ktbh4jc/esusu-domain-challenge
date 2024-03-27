package main

import (
	"encoding/json"
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

func (m *meme) makeMap() map[string]string {
	return map[string]string{
		"top_text":       m.TopText,
		"bottom_text":    m.BottomText,
		"image_location": m.ImageLocation,
	}
}

func TestMemeRoute(t *testing.T) {
	body := defaultMeme.makeMap()

	router := setupRouter()

	recorder := performRequest(router, "GET", "/memes")

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, body, response)
}
