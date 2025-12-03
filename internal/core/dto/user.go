package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/xkarasb/blog/pkg/types"
)

// easyjson:skip
//
//	@Description	UserDB represent user from data base
type UserDB struct {
	UserId                 uuid.UUID  `json:"user_id" db:"user_id"`
	Email                  string     `json:"email" db:"email"`
	PasswordHash           string     `db:"password_hash"`
	Role                   types.Role `json:"role" db:"role"`
	RefreshToken           string     `json:"refresh_token" db:"refresh_token"`
	RefreshTokenExpiryTime time.Time  `db:"refresh_token_expiry_time"`
} //	@name	UserDB
