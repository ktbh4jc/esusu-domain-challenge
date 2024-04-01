package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"maas/loggers"
	meme_maker "maas/meme-maker"
	query_params "maas/query-params"
	user_db "maas/user-db"
	user_service "maas/user-service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

func setupUserService(client *mongo.Client, ctx *context.Context) *user_service.UserService {
	mongoUserDb := user_db.NewMongoDBUserRepository(client, ctx)
	return user_service.NewUserService(mongoUserDb)
}

func connect(ctx context.Context) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	mongoUriTemplate := os.Getenv("MONGO_URI_TEMPLATE")
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoCluster := os.Getenv("MONGO_CLUSTER")
	MongoAppName := os.Getenv("MONGO_APP_NAME")

	mongoUri := fmt.Sprintf(
		mongoUriTemplate,
		mongoUsername,
		mongoPassword,
		mongoCluster,
		MongoAppName)
	fmt.Println(mongoUri)

	opts := options.Client().ApplyURI(mongoUri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func setupRouter(userService *user_service.UserService) *gin.Engine {
	router := gin.Default()
	router.GET("/memes", getMeme)
	router.GET("/mongo", userService.Ping)
	router.POST("/users/reset", userService.ResetDb)
	router.GET("/users/debug", userService.AllUsersDebug)
	router.GET("/users/:id", userService.UserById)
	return router
}

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		loggers.ErrorLog.Println("Error loading .env file")
		os.Exit(1)
	}

	loggers.Init()
	ctx := context.Background()
	client, err := connect(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	userService := setupUserService(client, &ctx)
	router := setupRouter(userService)

	rootURL := os.Getenv("ROOT_URL")
	router.Run(rootURL)

}
