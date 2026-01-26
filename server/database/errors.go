package database

import "errors"

var (
	ErrQuoteNotFound     = errors.New("quote not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrUsernameExists    = errors.New("username already exists")
	ErrEmailExists       = errors.New("email already exists")
)
