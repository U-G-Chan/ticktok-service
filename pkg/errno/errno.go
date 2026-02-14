package errno

import "fmt"

type Errno struct {
	Code    int
	Message string
}

func (e Errno) Error() string {
	return e.Message
}

// Err represents an error
type Err struct {
	Code    int
	Message string
	Err     error
}

func NewErrNo(code int, message string) Errno {
	return Errno{
		Code:    code,
		Message: message,
	}
}

func (err Err) Error() string {
	return fmt.Sprintf("Err - code: %d, message: %s, error: %s", err.Code, err.Message, err.Err)
}

func DecodeErr(err error) (int, string) {
	if err == nil {
		return Success.Code, Success.Message
	}

	switch typed := err.(type) {
	case *Err:
		return typed.Code, typed.Message
	case *Errno:
		return typed.Code, typed.Message
	}

	return InternalServerError.Code, err.Error()
}

var (
	Success             = NewErrNo(0, "Success")
	InternalServerError = NewErrNo(10001, "Internal server error")
	ErrBind             = NewErrNo(10002, "Error occurred while binding the request body to the struct.")
	ErrValidation       = NewErrNo(10003, "Validation failed.")
	ErrDatabase         = NewErrNo(10004, "Database error.")
	ErrToken            = NewErrNo(10005, "Error occurred while signing the JSON web token.")

	// User errors
	ErrUserNotFound      = NewErrNo(20001, "The user was not found.")
	ErrPasswordIncorrect = NewErrNo(20002, "The password is incorrect.")
)
