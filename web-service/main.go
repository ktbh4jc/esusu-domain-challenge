package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type meme struct {
	TopText       string `json:"top_text"`
	BottomText    string `json:"bottom_text"`
	ImageLocation string `json:"image_location"`
}

var defaultMeme = meme{TopText: "Up Top", BottomText: "Bottom Text", ImageLocation: "Nowhere and everywhere"}

func getMeme(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, defaultMeme)
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
