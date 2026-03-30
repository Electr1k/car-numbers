package repository

import "errors"

var (
	ErrConnectionFailed = errors.New("database connection failed")
	ErrQueryFailed      = errors.New("database query failed")
	ErrRecordNotFound   = errors.New("record not found")
)
