package model

import (
	"errors"
)

var (
	ErrURLNotFound      = errors.New("URL not found")
	ErrURLAlreadyExists = errors.New("URL already exists")
	ErrDeleted          = errors.New("URL was deleted")
)
