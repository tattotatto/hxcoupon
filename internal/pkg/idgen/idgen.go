package idgen

import "github.com/google/uuid"

func New() string {
	return uuid.New().String()
}

func Short() string {
	return uuid.New().String()[:8]
}
