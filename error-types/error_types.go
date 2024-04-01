package error_types

import "fmt"

type MongoConnectionError struct {
	Err error
}

func (m *MongoConnectionError) Error() string {
	return fmt.Sprintf("Unable to connect to MongoDB. Returned error:\n%s", m.Err.Error())
}

type BadEnvironmentError struct {
	Err error
}

func (b *BadEnvironmentError) Error() string {
	return fmt.Sprintf("Bad Environment Error:\n%s", b.Err.Error())
}
