package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

type MockReaderService struct {
	mock.Mock
}

func (m *MockReaderService) NewPost(authorId uuid.UUID, post *dto.CreatePostRequest) (*dto.CreatePostResponse, error) {
	args := m.Called(authorId, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CreatePostResponse), args.Error(1)
}

func (m *MockReaderService) GetPublishedPosts() ([]*dto.GetPostResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.GetPostResponse), args.Error(1)
}

func (m *MockReaderService) GetAuthorPosts(authorId uuid.UUID) ([]*dto.GetPostResponse, error) {
	args := m.Called(authorId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.GetPostResponse), args.Error(1)
}

func TestReaderController_CreatePostHandler(t *testing.T) {
	userId := uuid.New()
	postId := uuid.New()
	user := &dto.UserDB{UserId: userId, Role: types.Author}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockReaderService)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name: "successful post creation",
			requestBody: dto.CreatePostRequest{
				IdempotencyKey: "key-123",
				Title:          "Test Post",
				Content:        "Test Content",
			},
			setupMock: func(m *MockReaderService) {
				m.On("NewPost", userId, mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(&dto.CreatePostResponse{
						PostId: postId,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.CreatePostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, postId, resp.PostId)
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "{invalid json}",
			setupMock:      func(m *MockReaderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectBody.Error())
			},
		},
		{
			name: "empty body",
			requestBody: dto.CreatePostRequest{
				IdempotencyKey: "",
				Title:          "",
				Content:        "",
			},
			setupMock: func(m *MockReaderService) {
				m.On("NewPost", userId, mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(&dto.CreatePostResponse{
						PostId: postId,
					}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name: "idempotency key already used",
			requestBody: dto.CreatePostRequest{
				IdempotencyKey: "used-key",
				Title:          "Test Post",
				Content:        "Test Content",
			},
			setupMock: func(m *MockReaderService) {
				m.On("NewPost", userId, mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(nil, errors.ErrorKeyIdempotencyAlreadyUsed)
			},
			expectedStatus: http.StatusConflict,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorKeyIdempotencyAlreadyUsed.Error())
			},
		},
		{
			name: "unexpected error",
			requestBody: dto.CreatePostRequest{
				IdempotencyKey: "key-123",
				Title:          "Test Post",
				Content:        "Test Content",
			},
			setupMock: func(m *MockReaderService) {
				m.On("NewPost", userId, mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(nil, errors.ErrorHttpNoAuth)
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
		{
			name:           "null body",
			requestBody:    "null",
			setupMock:      func(m *MockReaderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:           "array instead of object",
			requestBody:    `[{"title": "Test"}]`,
			setupMock:      func(m *MockReaderService) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReaderService{}
			tt.setupMock(mockService)

			controller := &ReaderController{service: mockService}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, user))

			rr := httptest.NewRecorder()
			controller.CreatePostHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "NewPost")
			}
		})
	}
}

func TestReaderController_ViewSelectionHandler(t *testing.T) {
	authorId := uuid.New()
	readerId := uuid.New()
	authorUser := &dto.UserDB{UserId: authorId, Role: types.Author}
	readerUser := &dto.UserDB{UserId: readerId, Role: types.Reader}
	invalidUser := &dto.UserDB{UserId: uuid.New(), Role: types.Role("invalid")}

	tests := []struct {
		name           string
		user           *dto.UserDB
		setupMock      func(*MockReaderService, uuid.UUID)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name: "author view - successful",
			user: authorUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetAuthorPosts", userId).
					Return([]*dto.GetPostResponse{
						{
							PostId: uuid.New(),
							Title:  "Author Post",
							Status: types.Draft,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp []*dto.GetPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp, 1)
				assert.Equal(t, "Author Post", resp[0].Title)
			},
		},
		{
			name: "reader view - successful",
			user: readerUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetPublishedPosts").
					Return([]*dto.GetPostResponse{
						{
							PostId: uuid.New(),
							Title:  "Published Post",
							Status: types.Published,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp []*dto.GetPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp, 1)
				assert.Equal(t, "Published Post", resp[0].Title)
			},
		},
		{
			name:           "invalid role",
			user:           invalidUser,
			setupMock:      func(m *MockReaderService, userId uuid.UUID) {},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectUser.Error())
			},
		},
		{
			name: "author view - service error",
			user: authorUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetAuthorPosts", userId).
					Return(nil, errors.ErrorHttpNoAuth)
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
		{
			name: "reader view - service error",
			user: readerUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetPublishedPosts").
					Return(nil, errors.ErrorHttpNoAuth)
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
		{
			name: "author view - empty posts",
			user: authorUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetAuthorPosts", userId).
					Return([]*dto.GetPostResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp []*dto.GetPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp, 0)
			},
		},
		{
			name: "reader view - empty posts",
			user: readerUser,
			setupMock: func(m *MockReaderService, userId uuid.UUID) {
				m.On("GetPublishedPosts").
					Return([]*dto.GetPostResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp []*dto.GetPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReaderService{}
			tt.setupMock(mockService, tt.user.UserId)

			controller := &ReaderController{service: mockService}

			req := httptest.NewRequest(http.MethodGet, "/posts", nil)
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, tt.user))

			rr := httptest.NewRecorder()
			controller.ViewSelectionHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestReaderController_ViewSelectionHandler_NoUser(t *testing.T) {
	mockService := &MockReaderService{}
	controller := &ReaderController{service: mockService}

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)

	rr := httptest.NewRecorder()
	controller.ViewSelectionHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "GetAuthorPosts")
	mockService.AssertNotCalled(t, "GetPublishedPosts")
}

func TestReaderController_CreatePostHandler_NoUser(t *testing.T) {
	mockService := &MockReaderService{}
	controller := &ReaderController{service: mockService}

	bodyBytes, _ := json.Marshal(dto.CreatePostRequest{Title: "Test", Content: "Content"})
	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.CreatePostHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "NewPost")
}
