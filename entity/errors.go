package entity

import "fmt"

var (
	ErrInvalidArguments = fmt.Errorf("invalid arguments")
	ErrEmptyCredentials = fmt.Errorf("empty username or password")
)
