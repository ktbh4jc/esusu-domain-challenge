package user_service

import (
	"maas/loggers"
	"net/http"

	error_types "maas/error-types"
	user_model "maas/user-model"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	ResetDb() ([]interface{}, error)
	Ping() error
	AllUsers() ([]user_model.User, error)
	User(id string) (*user_model.User, error)
}

type UserService struct {
	Repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

func (s *UserService) Ping(ginContext *gin.Context) {
	// Send a ping to confirm a successful connection
	if err := s.Repo.Ping(); err != nil {
		loggers.ErrorLog.Printf("Error pinging database:\n%s", err.Error())
		ginContext.IndentedJSON(http.StatusInternalServerError, "error pinging database")
		return
	}

	loggers.InfoLog.Printf("Pinged your deployment. You successfully connected to MongoDB!")
	ginContext.IndentedJSON(http.StatusOK, "Connection good")
}

// Resets the database to the values in default_db_data.DefaultUsers
// Returns a 400 if attempted from a non-local environment
// Returns a 500 if another error is raised
// Otherwise returns 200 and the generated object ids
func (s *UserService) ResetDb(ginContext *gin.Context) {
	// Tell the user_db to reset the DB
	userIds, err := s.Repo.ResetDb()
	if err != nil {
		switch err.(type) {
		default:
			loggers.ErrorLog.Printf("Encountered an error: %s", err.Error())
			ginContext.IndentedJSON(http.StatusInternalServerError, "error with DB reset")
		case *error_types.BadEnvironmentError:
			loggers.ErrorLog.Print(err.Error())
			ginContext.IndentedJSON(http.StatusBadRequest, "Unable to reset DB in current environment")
		}
		return
	}
	ginContext.IndentedJSON(http.StatusOK, userIds)
}

// GETs all users without verifying requesting user is an admin
// Meant as a debugging tool for graders since the users are on a cloud instance
func (s *UserService) AllUsersDebug(ginContext *gin.Context) {
	users, err := s.Repo.AllUsers()
	if err != nil {
		loggers.ErrorLog.Printf("Error getting users:\n%s", err.Error())
		ginContext.IndentedJSON(http.StatusInternalServerError, "error getting users")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, users)
}

func (s *UserService) UserById(ginContext *gin.Context) {
	user, err := s.Repo.User(ginContext.Param("id"))
	if err != nil {
		loggers.ErrorLog.Printf("Error getting user:\n%s", err.Error())
		ginContext.IndentedJSON(http.StatusNotFound, "error getting user")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, user)
}
