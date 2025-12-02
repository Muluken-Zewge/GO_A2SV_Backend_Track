package domain

import "errors"

// global,reusable domain errors
var ErrNotFound = errors.New("resource not found")
var ErrAleadyExists = errors.New("resource already exists")
var ErrValidation = errors.New("input validation failed")
var ErrInvalidCredential = errors.New("invalid username or password")
