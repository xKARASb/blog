package handlers

import (
	"bytes"
	"encoding/json"
	gerrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
)

type MockAuthService struct {
	mock.Mock
	secret string
}

func (m *MockAuthService) RegistrateUser(user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RegistrateUserResponse), args.Error(1)
}

func (m *MockAuthService) LoginUser(user *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginUserResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(user *dto.RefreshRequest) (*dto.RefreshResponse, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RefreshResponse), args.Error(1)
}

func (m *MockAuthService) AuthorizeUser(token string) (*dto.UserDB, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDB), args.Error(1)
}

func TestAuthController_RegisterHandler(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkBody      func(*testing.T, string)
	}{
		{
			name: "successful registration",
			requestBody: dto.RegistrateUserRequest{
				Email:    "test@example.com",
				Password: "Password123!",
				Role:     types.Author,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(&dto.RegistrateUserResponse{
						Id:           id,
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var resp dto.RegistrateUserResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, id, resp.Id)
				assert.NotEmpty(t, resp.AccessToken)
			},
		},
		{
			name: "user already exists",
			requestBody: dto.RegistrateUserRequest{
				Email:    "existing@example.com",
				Password: "Password123!",
				Role:     types.Author,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrorRepositoryUserAlreadyExsist)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "bad email",
			requestBody: dto.RegistrateUserRequest{
				Email:    "bad email",
				Password: "Password123!",
				Role:     types.Author,
			},
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad role",
			requestBody: dto.RegistrateUserRequest{
				Email:    "existing@example.com",
				Password: "Password123!",
				Role:     types.Role("BadRole"),
			},
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad body",
			requestBody: dto.RegistrateUserRequest{
				Password: "Password123!",
				Role:     types.Author,
			},
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unexpected error",
			requestBody: dto.RegistrateUserRequest{
				Email:    "existing@example.com",
				Password: "Password123!",
				Role:     types.Author,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrorHttpNoAuth)
			},
			expectedStatus: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{secret: "test-secret"}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}
			controller := &AuthController{service: mockService}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			controller.RegisterHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
func TestAuthController_LoginHandler(t *testing.T) {
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	refreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name: "successful login",
			requestBody: dto.LoginUserRequest{
				Email:    "user@example.com",
				Password: "Password123!",
			},
			setupMock: func(m *MockAuthService) {
				m.On("LoginUser", mock.MatchedBy(func(req *dto.LoginUserRequest) bool {
					return req.Email == "user@example.com" && req.Password == "Password123!"
				})).Return(&dto.LoginUserResponse{
					AccessToken:  accessToken,
					RefreshToken: refreshToken,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.LoginUserResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, accessToken, resp.AccessToken)
				assert.Equal(t, refreshToken, resp.RefreshToken)
			},
		},

		{
			name:           "invalid JSON",
			requestBody:    "{invalid json}",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectBody.Error())
			},
		},
		{
			name: "missing email",
			requestBody: dto.LoginUserRequest{
				Password: "Password123!",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "missing password",
			requestBody: dto.LoginUserRequest{
				Email: "user@example.com",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "invalid email format",
			requestBody: dto.LoginUserRequest{
				Email:    "not-an-email",
				Password: "Password123!",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "email")
			},
		},
		{
			name: "empty email",
			requestBody: dto.LoginUserRequest{
				Email:    "",
				Password: "Password123!",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "empty password",
			requestBody: dto.LoginUserRequest{
				Email:    "user@example.com",
				Password: "",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},

		{
			name: "user not found",
			requestBody: dto.LoginUserRequest{
				Email:    "nonexistent@example.com",
				Password: "Password123!",
			},
			setupMock: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrorRepositoryEmailNotExsist)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorRepositoryEmailNotExsist.Error())
			},
		},
		{
			name: "wrong password",
			requestBody: dto.LoginUserRequest{
				Email:    "user@example.com",
				Password: "WrongPassword!",
			},
			setupMock: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrorRepositoryEmailNotExsist)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorRepositoryEmailNotExsist.Error())
			},
		},
		{
			name: "jwt generation error",
			requestBody: dto.LoginUserRequest{
				Email:    "user@example.com",
				Password: "Password123!",
			},
			setupMock: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
		{
			name:        "malformed JSON with extra fields",
			requestBody: `{"email": "user@example.com", "password": "Password123!", "extra": "field"}`,
			setupMock: func(m *MockAuthService) {
				m.On("LoginUser", mock.MatchedBy(func(req *dto.LoginUserRequest) bool {
					return req.Email == "user@example.com" && req.Password == "Password123!"
				})).Return(&dto.LoginUserResponse{
					AccessToken:  accessToken,
					RefreshToken: refreshToken,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
		},
		{
			name:           "null body",
			requestBody:    "null",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "empty object",
			requestBody:    "{}",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "array instead of object",
			requestBody:    `[{"email": "user@example.com", "password": "Password123!"}]`,
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{}
			tt.setupMock(mockService)

			controller := &AuthController{service: mockService}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err, "Failed to marshal request body")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
				bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			controller.LoginHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)

			} else {
				mockService.AssertNotCalled(t, "LoginUser")

				if tt.expectedStatus == http.StatusBadRequest {
					assert.NotEmpty(t, rr.Body.String(),
						"Validation error should have a message")
				}
			}
		})
	}
}
func TestAuthController_RefreshHandler(t *testing.T) {
	validRefreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJleHAiOjE2OTg3NjUyMDB9.signature"
	newAccessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE2OTg3NjUyMDB9.new_signature"
	// newRefreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJleHAiOjE5OTg3NjUyMDB9.new_signature"

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name: "successful token refresh",
			requestBody: dto.RefreshRequest{
				RefreshToken: validRefreshToken,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.MatchedBy(func(req *dto.RefreshRequest) bool {
					return req.RefreshToken == validRefreshToken
				})).Return(&dto.RefreshResponse{
					AccessToken: newAccessToken,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.RefreshResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, newAccessToken, resp.AccessToken)
				assert.NotEmpty(t, resp.AccessToken)
			},
		},

		{
			name:           "invalid JSON",
			requestBody:    "{invalid json}",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectBody.Error())
			},
		},
		{
			name: "empty refresh token",
			requestBody: dto.RefreshRequest{
				RefreshToken: "",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
			},
		},
		{
			name:           "missing refresh_token field",
			requestBody:    map[string]interface{}{},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "null body",
			requestBody:    "null",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "empty object",
			requestBody:    "{}",
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "array instead of object",
			requestBody:    `[{"refresh_token": "token"}]`,
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "refresh_token is not a string",
			requestBody: map[string]interface{}{
				"refresh_token": 12345,
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "invalid token error",
			requestBody: dto.RefreshRequest{
				RefreshToken: "invalid.token.here",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, errors.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpBadRefresh.Error())
				assert.NotContains(t, body, errors.ErrorInvalidToken.Error())
			},
		},
		{
			name: "expired refresh token",
			requestBody: dto.RefreshRequest{
				RefreshToken: "expired.token.here",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, errors.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpBadRefresh.Error())
			},
		},
		{
			name: "malformed token",
			requestBody: dto.RefreshRequest{
				RefreshToken: "not.a.jwt.token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, errors.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
		},
		{
			name: "user not found for token",
			requestBody: dto.RefreshRequest{
				RefreshToken: validRefreshToken,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, errors.ErrorHttpIncorrectUser)
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectUser.Error())
			},
		},
		{
			name: "database error during refresh",
			requestBody: dto.RefreshRequest{
				RefreshToken: validRefreshToken,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, gerrors.New("database connection failed"))
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "database connection failed")
			},
		},
		{
			name: "jwt generation error",
			requestBody: dto.RefreshRequest{
				RefreshToken: validRefreshToken,
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(nil, errors.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
		},
		{
			name: "token with special characters",
			requestBody: dto.RefreshRequest{
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.AnythingOfType("*dto.RefreshRequest")).
					Return(&dto.RefreshResponse{
						AccessToken: newAccessToken,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{}
			tt.setupMock(mockService)

			controller := &AuthController{service: mockService}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err, "Failed to marshal request body")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token",
				bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			controller.RefreshHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)

			} else {
				mockService.AssertNotCalled(t, "RefreshUser")

				if tt.expectedStatus == http.StatusBadRequest {
					assert.NotEmpty(t, rr.Body.String(),
						"Validation error should have a message")
				}
			}
		})
	}
}
