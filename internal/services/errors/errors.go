package svcErr

import (
	"fmt"
)

type NotFoundError struct {
	Entity string
	Field  string
	Value  string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with %s '%s' not found", e.Entity, e.Field, e.Value)
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
