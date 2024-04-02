package user_service

import (
	auth_service "maas/auth-service"
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
	Auth auth_service.AuthService
}

func NewUserService(repo UserRepository, auth auth_service.AuthService) *UserService {
	return &UserService{
		Repo: repo,
		Auth: auth,
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
	err := s.requireCallerOrAdmin(ginContext)
	if err != nil {
		return
	}

	user, err := s.Repo.User(ginContext.Param("id"))
	if err != nil {
		loggers.ErrorLog.Printf("Error getting user:\n%s", err.Error())
		ginContext.IndentedJSON(http.StatusNotFound, "error getting user")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, user)
}

// GETs all users, requires requesting user to be admin
func (s *UserService) AllUsers(ginContext *gin.Context) {
	err := s.requireAdmin(ginContext)
	if err != nil {
		return
	}

	users, err := s.Repo.AllUsers()
	if err != nil {
		loggers.ErrorLog.Printf("Error getting users:\n%s", err.Error())
		ginContext.IndentedJSON(http.StatusInternalServerError, "error getting users")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, users)
}

// RequireAdmin
func (s *UserService) requireAdmin(ginContext *gin.Context) error {
	authHeader := ginContext.Request.Header.Get("auth")
	isAdmin, err := s.Auth.IsAdmin(authHeader)
	if err != nil {
		authResponse(err, ginContext)
		return err
	}
	if isAdmin {
		return nil
	}
	err = &error_types.NotAdminError{}
	authResponse(err, ginContext)
	return err
}

func (s *UserService) requireCallerOrAdmin(ginContext *gin.Context) error {
	authHeader := ginContext.Request.Header.Get("auth")
	isAdmin, err := s.Auth.IsCallerOrAdmin(authHeader, ginContext.Param("id"))
	if err != nil {
		authResponse(err, ginContext)
		return err
	}
	if isAdmin {
		return nil
	}
	err = &error_types.NoAccessError{}
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
	case *error_types.NoAccessError:
		loggers.ErrorLog.Print(err.Error())
		ginContext.IndentedJSON(http.StatusForbidden, "forbidden")
	}
}
