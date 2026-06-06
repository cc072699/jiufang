package errors

import "errors"

var (
	ErrOldPasswordIncorrect    = errors.New("old password is incorrect")
	ErrPasswordTooShort        = errors.New("password length must be between 6 and 20 characters")
	ErrPasswordTooLong         = errors.New("password length must be between 6 and 20 characters")
	ErrPasswordNotMatch        = errors.New("new password and confirm password do not match")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("username already exists")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidToken            = errors.New("invalid token")
	ErrGroupNotFound           = errors.New("user group not found")
	ErrGroupNameExists         = errors.New("user group name already exists")
	ErrPresetGroupCannotDelete = errors.New("preset user group cannot be deleted")
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrAccountNotFound         = errors.New("account not found")
	ErrInvalidCredentials      = errors.New("invalid username or password")
	ErrUserDisabled            = errors.New("user is disabled")
	ErrTokenExpired            = errors.New("token is expired")
	ErrTokenInvalid            = errors.New("token is invalid")
	ErrInvalidRequest          = errors.New("invalid request")
)

const (
	CodeSuccess       = 200
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeConflict      = 409
	CodeInternalError = 500
)

const (
	CodeOldPasswordIncorrect    = 40001
	CodePasswordTooShort        = 40002
	CodePasswordTooLong         = 40003
	CodePasswordNotMatch        = 40004
	CodeUserNotFound            = 40401
	CodeUserAlreadyExists       = 40902
	CodeEmailAlreadyExists      = 40903
	CodeInvalidToken            = 40101
	CodeGroupNotFound           = 40402
	CodeGroupNameExists         = 40901
	CodePresetGroupCannotDelete = 40302
	CodePermissionNotFound      = 40403
	CodeAccountNotFound         = 40102
	CodeInvalidCredentials      = 40103
	CodeUserDisabled            = 40301
	CodeTokenExpired            = 40104
	CodeTokenInvalid            = 40105
	CodeInvalidRequest          = 40005
)
