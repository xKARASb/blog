package errors

import "errors"

var (
	ErrorRepositoryUserAlreadyExsist = errors.New("user already exsist")
	ErrorServiceEmailInvalid         = errors.New("invalid email")
	ErrorRepositoryEmailNotExsist    = errors.New("email not exsist")
	ErrorRepositoryBadRole           = errors.New("bad role")
	ErrorInvalidToken                = errors.New("invalid token")
	ErrorKeyIdempotencyAlreadyUsed   = errors.New("key idempotency already used")
	ErrorServiceNoAccess             = errors.New("no access to content")
	ErrorServiceIncorrectData        = errors.New("incorrect data")
	ErrorHttpIncorrectUser           = errors.New("incorrect user")
	ErrorHttpNoAuth                  = errors.New("no authorization provided")
	ErrorHttpIncorrectBody           = errors.New("incorrect body")
	ErrorHttpIncorrectEmail          = errors.New("email or password incorrect")
	ErrorHttpBadRefresh              = errors.New("refresh token expired or incorrect")
	ErrorHttpPostNotFound            = errors.New("post not found")
	ErrorHttpImageNotFound           = errors.New("image not found")
	ErrorHttpAccessDenied            = errors.New("access denied")
	ErrorHttpIncorrectStatus         = errors.New("incorrect status")
)
