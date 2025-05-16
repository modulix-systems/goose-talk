package storage

import "errors"

var (
	ErrNotFound      = errors.New("Record not found")
	ErrAlreadyExists = errors.New("Record already exists")
)
