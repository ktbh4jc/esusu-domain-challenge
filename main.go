package main

import (
	"context"
	"fmt"
	"log"
	meme_maker "maas/meme-maker"
	query_params "maas/query-params"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getMeme(ctx *gin.Context) {
	queryParams, err := query_params.ExtractParams(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}
	ctx.IndentedJSON(http.StatusOK, meme_maker.BuildMeme(queryParams))
}

func pingMongo(ctx *gin.Context) {
	{
		// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		mongoUsername := os.Getenv("MONGO_USERNAME")
		mongoPassword := os.Getenv("MONGO_PASSWORD")
		mongoCluster := os.Getenv("MONGO_CLUSTER")
		MongoAppName := os.Getenv("MONGO_APP_NAME")

		mongoUri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=%s", mongoUsername, mongoPassword, mongoCluster, MongoAppName)

		opts := options.Client().ApplyURI(mongoUri).SetServerAPIOptions(serverAPI)

		// Create a new client and connect to the server
		client, err := mongo.Connect(context.TODO(), opts)
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, "error connecting to mongo db")
			panic(err)
		}

		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()

		// Send a ping to confirm a successful connection
		if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, "error connecting to mongo db")
			panic(err)
		}
		fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
		ctx.IndentedJSON(http.StatusOK, "Connection good")
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/memes", getMeme)
	router.GET("/mongo", pingMongo)
	return router
}

func main() {
	router := setupRouter()
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rootURL := os.Getenv("ROOT_URL")
	router.Run(rootURL)
}
