package user_service

import (
	// "encoding/json"

	"fmt"
	auth_service "maas/auth-service"
	"maas/loggers"
	"net/http"
	"strconv"

	error_types "maas/error-types"
	"maas/models"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	ResetDb() ([]interface{}, error)
	Ping() error
	AllUsers() ([]models.User, error)
	User(id string) (*models.User, error)
	NewUser(user models.User) (interface{}, error)
	UpdateUser(id string, user *models.User) error
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

// GETs a user by ID. Needs to be either the requesting user getting their own info or an admin
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

// POST for a new user. Only an admin can create a new user.
func (s *UserService) NewUser(ginContext *gin.Context) {
	err := s.requireAdmin(ginContext)
	if err != nil {
		return
	}

	user, err := s.userFromGinContext(ginContext)
	if err != nil {
		// error responses set in userFromGinContext
		return
	}
	err = s.ensureAuthKeyIsNew(*user, ginContext)
	if err != nil {
		// error responses set in ensureAuthKeyIsNew
		return
	}

	result, err := s.Repo.NewUser(*user)
	if err != nil {
		loggers.ErrorLog.Printf("Error encountered creating user: %s", err)
		ginContext.IndentedJSON(http.StatusBadRequest, "Encountered error creating new user")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, result)
}

// PATCH an existing user. Only an admin can do this.
func (s *UserService) UpdateUser(ginContext *gin.Context) {
	id := ginContext.Param("id")

	err := s.requireAdmin(ginContext)
	if err != nil {
		return
	}

	// Confirming user exists, feels like this could be combined with the update
	_, err = s.Repo.User(id)
	if err != nil {
		loggers.ErrorLog.Printf("Encountered error getting user: %s%v", id, err)
		ginContext.IndentedJSON(http.StatusNotFound, "Unable to find that user")
		return
	}

	//Build user from form provided context
	newUser, err := s.userFromGinContext(ginContext)
	if err != nil {
		return
	}

	err = s.Repo.UpdateUser(id, newUser)
	if err != nil {
		ginContext.IndentedJSON(http.StatusInternalServerError, "There was an error, please try again later")
		return
	}
	ginContext.IndentedJSON(http.StatusOK, "successfully updated user")
}

// Auth related helpers
func (s *UserService) userFromGinContext(ginContext *gin.Context) (*models.User, error) {
	tokens, err := strconv.Atoi(ginContext.PostForm("tokens_remaining"))
	if err != nil {
		loggers.ErrorLog.Printf("Error encountered creating user: %s", err)
		ginContext.IndentedJSON(http.StatusBadRequest, "tokens_remaining must be an int")
		return nil, err
	}
	isAdmin, err := strconv.ParseBool(ginContext.PostForm("is_admin"))
	if err != nil {
		loggers.ErrorLog.Printf("Error encountered creating user: %s", err)
		ginContext.IndentedJSON(http.StatusBadRequest, "is_admin must be an bool")
		return nil, err
	}

	user := &models.User{
		UserId:          ginContext.PostForm("user_id"),
		TokensRemaining: tokens,
		AuthKey:         ginContext.PostForm("auth_key"),
		IsAdmin:         isAdmin,
	}
	return user, nil
}

func (s *UserService) ensureAuthKeyIsNew(user models.User, ginContext *gin.Context) error {
	userResult, err := s.Auth.Repo.UserByAuthHeader(user.AuthKey)
	if err != nil {
		switch err.(type) {
		default:
			ginContext.IndentedJSON(http.StatusInternalServerError, "issue confirming provided auth header is unused")
			return err
		case *error_types.UnableToLocateDocumentError:
			return nil
		}
	}
	// Always fun to have an excuse for an error message that you would never want to use IRL
	ginContext.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("User %s is already using that auth header", userResult.ID.Hex()))
	return &error_types.AuthKeyAlreadyTakenError{}
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
