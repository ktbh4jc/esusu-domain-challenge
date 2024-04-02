package meme_service

import (
	auth_service "maas/auth-service"
	error_types "maas/error-types"
	"maas/loggers"
	"maas/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Lon   float64 `json:"lon"`
	Lat   float64 `json:"lat"`
	Query string  `json:"query"`
}

type UserRepository interface {
	User(id string) (*models.User, error)
	UserByAuthHeader(auth string) (*models.User, error)
	UpdateUser(id string, user *models.User) error
}

type MemeProvider interface {
	BuildMeme(*QueryParams) (*models.Meme, error)
}

type MemeService struct {
	UserRepo     UserRepository
	Auth         auth_service.AuthService
	MemeProvider MemeProvider
}

func NewMemeService(userRepo UserRepository, auth auth_service.AuthService, memeProvider MemeProvider) *MemeService {
	return &MemeService{
		UserRepo:     userRepo,
		Auth:         auth,
		MemeProvider: memeProvider,
	}
}

func (s *MemeService) ExtractParams(c *gin.Context) (*QueryParams, error) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		if c.Query("lat") == "" {
			lat = 0
		} else {
			return nil, err
		}
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		if c.Query("lon") == "" {
			lon = 0
		} else {
			return nil, err
		}
	}
	queryParams := &QueryParams{
		Lat:   lat,
		Lon:   lon,
		Query: c.Query("query"),
	}
	return queryParams, nil
}

func (s *MemeService) GetMeme(ginContext *gin.Context) {
	err := s.requireAuthenticated(ginContext)
	if err != nil {
		return
	}
	params, err := s.ExtractParams(ginContext)
	if err != nil {
		loggers.ErrorLog.Printf("Encountered an error making a meme%s\n", err)
		ginContext.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}

	user, err := s.UserRepo.UserByAuthHeader(ginContext.Request.Header.Get("auth"))
	if err != nil {
		loggers.ErrorLog.Printf("Encountered an error making a meme%s\n", err)
		ginContext.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}

	if user.TokensRemaining < 1 {
		loggers.ErrorLog.Printf("Encountered an error making a meme%s\n", err)
		ginContext.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "Tokens needed to make more memes. Buy some!"})
		return
	}

	user.TokensRemaining -= 1
	err = s.UserRepo.UpdateUser(user.ID.Hex(), user)
	if err != nil {
		loggers.ErrorLog.Printf("Encountered an error making a meme%s\n", err)
		ginContext.IndentedJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
		return
	}

	meme, err := s.MemeProvider.BuildMeme(params)
	if err != nil {
		loggers.ErrorLog.Printf("Encountered an error making a meme%s\n", err)
		ginContext.IndentedJSON(http.StatusInternalServerError, map[string]string{"error": "Unable to make meme"})
		return
	}

	ginContext.IndentedJSON(http.StatusOK, meme)
}

func (s *MemeService) requireAuthenticated(ginContext *gin.Context) error {
	authHeader := ginContext.Request.Header.Get("auth")
	IsAuthenticated, err := s.Auth.IsAuthenticated(authHeader)
	if err != nil {
		authResponse(err, ginContext)
		return err
	}
	if IsAuthenticated {
		return nil
	}
	err = &error_types.UserNotFoundError{}
	authResponse(err, ginContext)
	return err
}
func authResponse(err error, ginContext *gin.Context) {
	switch err.(type) {
	default:
		loggers.ErrorLog.Printf("Encountered an error during authentication: %s", err.Error())
		ginContext.IndentedJSON(http.StatusForbidden, "forbidden")
	case *error_types.NoAuthHeaderError:
		loggers.ErrorLog.Print(err.Error())
		ginContext.IndentedJSON(http.StatusUnauthorized, "unauthorized")
	case *error_types.AuthUserNotFoundError:
		loggers.ErrorLog.Print(err.Error())
		ginContext.IndentedJSON(http.StatusForbidden, "forbidden")
	}
}
