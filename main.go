package main

import (
	"github.com/gin-gonic/gin"
	meme_maker "maas/meme-maker"
	query_params "maas/query-params"
	"net/http"
)

func getMeme(context *gin.Context) {
	queryParams, err := query_params.ExtractParams(context)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	context.IndentedJSON(http.StatusOK, meme_maker.BuildMeme(queryParams))
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/memes", getMeme)
	return router
}

func main() {
	router := setupRouter()
	router.Run("localhost:8080")
}
