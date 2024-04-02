package error_types

import "fmt"

type MongoConnectionError struct {
	Err error
}

func (e *MongoConnectionError) Error() string {
	return fmt.Sprintf("Unable to connect to MongoDB. Returned error:\n%s", e.Err.Error())
}

type BadEnvironmentError struct {
	Err error
}

func (e *BadEnvironmentError) Error() string {
	return fmt.Sprintf("Bad Environment Error:\n%s", e.Err.Error())
}

type NoAuthHeaderError struct{}

func (e *NoAuthHeaderError) Error() string {
	return "No Auth Header on request"
}

type NotAdminError struct{}

func (e *NotAdminError) Error() string {
	return "Requesting user is not an admin"
}

type AuthUserNotFoundError struct{}

func (e *AuthUserNotFoundError) Error() string {
	return "User could not be authenticated"
}

type UserNotFoundError struct{}

func (e *UserNotFoundError) Error() string {
	return "User not found"
}

type NoAccessError struct{}

func (e *NoAccessError) Error() string {
	return "User does not have access"
}

type AuthKeyAlreadyTakenError struct{}

func (e *AuthKeyAlreadyTakenError) Error() string {
	return "A user is already using that auth key"
}

type UnableToLocateDocumentError struct {
	Err error
}

func (e *UnableToLocateDocumentError) Error() string {
	return fmt.Sprintf("Unable to locate document:\n%s", e.Err.Error())
}
