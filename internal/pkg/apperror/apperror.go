package apperror

import "hxcoupon/internal/pkg/errcode"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(code int) *AppError {
	return &AppError{Code: code, Message: errcode.Message(code)}
}

func NewWithErr(code int, err error) *AppError {
	return &AppError{Code: code, Message: errcode.Message(code), Err: err}
}

func NewWithMsg(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}
