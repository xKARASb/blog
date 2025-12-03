package dto

import "github.com/google/uuid"

// @Description	Request payload for registering a new user
type RegistrateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
} //	@name	UserRegistrationRequest

// @Description	Response with authentication tokens after registration
type RegistrateUserResponse struct {
	Id           uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
} //	@name	UserRegistrationResponse

// @Description	Request payload for user authentication
type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} //	@name	UserLoginRequest

// @Description	Response with authentication tokens after login
type LoginUserResponse struct {
	Id           uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
} //	@name	UserLoginResponse

// @Description	Request to refresh access token using refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
} //	@name	TokenRefreshRequest

// @Description	Response with new access token
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
} //	@name	TokenRefreshResponse
