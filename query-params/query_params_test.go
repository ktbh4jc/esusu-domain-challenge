package query_params

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func buildTestContext(path string) *gin.Context {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()

	context, _ := gin.CreateTestContext(recorder)
	context.Request, _ = http.NewRequest("GET", path, nil)
	return context
}

func TestExtractParams_WithNoParams_ReturnsZeroValueParams(t *testing.T) {
	path := "/test"

	context := buildTestContext(path)
	params, err := ExtractParams(context)
	assert.Nil(t, err)
	assert.Equal(t, &QueryParams{}, params)
}

func TestExtractParams_WithParams_SetsParams(t *testing.T) {
	path := "/test?query=test&lat=1&lon=2"

	context := buildTestContext(path)
	params, err := ExtractParams(context)
	assert.Nil(t, err)
	assert.Equal(t, &QueryParams{Query: "test", Lat: 1, Lon: 2}, params)
}

func TestExtractParams_WithBadLon_ThrowsError(t *testing.T) {
	path := "/test?lon=five"

	context := buildTestContext(path)
	params, err := ExtractParams(context)
	assert.Error(t, err)
	assert.Nil(t, params)
}

func TestExtractParams_WithBadLat_ThrowsError(t *testing.T) {
	path := "/test?lat=lon"

	context := buildTestContext(path)
	params, err := ExtractParams(context)
	assert.Error(t, err)
	assert.Nil(t, params)
}
