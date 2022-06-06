package models

import (
	"fmt"
)

var (
	internalServerError error
	loginFailError      error
)

func InitError() {
	internalServerError = _e("Internal server error, try again later please")
	loginFailError = _e("Login fail, check your username and password")
}

func _e(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}
