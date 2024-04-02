package main

import (
	"context"
	"fmt"

	// "net/http"
	"os"

	auth_service "maas/auth-service"
	"maas/loggers"
	meme_maker "maas/meme-maker"
	meme_service "maas/meme-service"

	// meme_maker "maas/meme-maker"
	// meme_service "maas/meme-service"
	user_db "maas/user-db"
	user_service "maas/user-service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupAuthAndUserServices(client *mongo.Client, ctx *context.Context) (*user_service.UserService, *auth_service.AuthService) {
	mongoUserDb := user_db.NewMongoDBUserRepository(client, ctx)
	authService := auth_service.NewAuthService(mongoUserDb)
	return user_service.NewUserService(mongoUserDb, *authService), *&authService
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

func setupRouter(userService *user_service.UserService, memeService *meme_service.MemeService) *gin.Engine {
	router := gin.Default()
	router.GET("/memes", memeService.GetMeme)
	router.GET("/mongo", userService.Ping)
	router.POST("/users/reset", userService.ResetDb)
	router.GET("/users/debug", userService.AllUsersDebug)
	router.GET("/users", userService.AllUsers)
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

	mongoUserDb := user_db.NewMongoDBUserRepository(client, &ctx)
	authService := auth_service.NewAuthService(mongoUserDb)
	userService := user_service.NewUserService(mongoUserDb, *authService)
	memeService := meme_service.NewMemeService(mongoUserDb, *authService, &meme_maker.MemeMaker{})
	router := setupRouter(userService, memeService)

	rootURL := os.Getenv("ROOT_URL")
	router.Run(rootURL)

}
