package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
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

type MockPosterService struct {
	mock.Mock
}

func (m *MockPosterService) EditPost(userId, postId uuid.UUID, post *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	args := m.Called(userId, postId, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EditPostResponse), args.Error(1)
}

func (m *MockPosterService) PublishPost(userId, postId uuid.UUID, post *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	args := m.Called(userId, postId, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublishPostResponse), args.Error(1)
}

func (m *MockPosterService) AddImage(userId, postId uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*dto.AddImageResponse, error) {
	args := m.Called(userId, postId, file, fileHeader)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AddImageResponse), args.Error(1)
}

func (m *MockPosterService) DeleteImage(userId, postId, imageId uuid.UUID) (*dto.DeleteImageResponse, error) {
	args := m.Called(userId, postId, imageId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DeleteImageResponse), args.Error(1)
}

func TestPosterController_EditPostHandler(t *testing.T) {
	userId := uuid.New()
	postId := uuid.New()
	user := &dto.UserDB{UserId: userId, Role: types.Author}

	tests := []struct {
		name           string
		postId         string
		requestBody    interface{}
		setupMock      func(*MockPosterService, uuid.UUID)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name:   "successful edit",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(&dto.EditPostResponse{
						PostId:         parsedPostId,
						AuthorId:       userId,
						IdempotencyKey: "key",
						Title:          "Updated Title",
						Content:        "Updated Content",
						Status:         types.Draft,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.EditPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, postId, resp.PostId)
				assert.Equal(t, "Updated Title", resp.Title)
			},
		},
		{
			name:           "invalid post ID",
			postId:         "invalid-uuid",
			requestBody:    dto.EditPostRequest{Title: "Title", Content: "Content"},
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpPostNotFound.Error())
			},
		},
		{
			name:           "invalid JSON",
			postId:         postId.String(),
			requestBody:    "{invalid json}",
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectBody.Error())
			},
		},
		{
			name:   "empty body",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "",
				Content: "",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(&dto.EditPostResponse{
						PostId:         parsedPostId,
						AuthorId:       userId,
						IdempotencyKey: "key",
						Title:          "",
						Content:        "",
						Status:         types.Draft,
					}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:   "no access",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "Title",
				Content: "Content",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrorServiceNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpAccessDenied.Error())
			},
		},
		{
			name:   "post not found",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "Title",
				Content: "Content",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpPostNotFound.Error())
			},
		},
		{
			name:   "incorrect data",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "Title",
				Content: "Content",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrorServiceIncorrectData)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, errors.ErrorHttpIncorrectStatus.Error())
			},
		},
		{
			name:   "unexpected error",
			postId: postId.String(),
			requestBody: dto.EditPostRequest{
				Title:   "Title",
				Content: "Content",
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("EditPost", userId, parsedPostId, mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPosterService{}
			parsedPostId, _ := uuid.Parse(tt.postId)
			tt.setupMock(mockService, parsedPostId)

			controller := &PosterController{service: mockService}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/post/%s", tt.postId), bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("postId", tt.postId)
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, user))

			rr := httptest.NewRecorder()
			controller.EditPostHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "EditPost")
			}
		})
	}
}

func TestPosterController_PublishHandler(t *testing.T) {
	userId := uuid.New()
	postId := uuid.New()
	user := &dto.UserDB{UserId: userId, Role: types.Author}

	tests := []struct {
		name           string
		postId         string
		requestBody    interface{}
		setupMock      func(*MockPosterService, uuid.UUID)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name:   "successful publish",
			postId: postId.String(),
			requestBody: dto.PublishPostRequest{
				Status: types.Published,
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("PublishPost", userId, parsedPostId, mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(&dto.PublishPostResponse{
						PostId: parsedPostId,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.PublishPostResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, postId, resp.PostId)
			},
		},
		{
			name:           "invalid post ID",
			postId:         "invalid-uuid",
			requestBody:    dto.PublishPostRequest{Status: types.Published},
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: false,
		},
		{
			name:           "invalid JSON",
			postId:         postId.String(),
			requestBody:    "{invalid json}",
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: false,
		},
		{
			name:   "no access",
			postId: postId.String(),
			requestBody: dto.PublishPostRequest{
				Status: types.Published,
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("PublishPost", userId, parsedPostId, mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrorServiceNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
		},
		{
			name:   "post not found",
			postId: postId.String(),
			requestBody: dto.PublishPostRequest{
				Status: types.Published,
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("PublishPost", userId, parsedPostId, mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: true,
		},
		{
			name:   "incorrect data",
			postId: postId.String(),
			requestBody: dto.PublishPostRequest{
				Status: types.Published,
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("PublishPost", userId, parsedPostId, mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrorServiceIncorrectData)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
		},
		{
			name:   "unexpected error",
			postId: postId.String(),
			requestBody: dto.PublishPostRequest{
				Status: types.Published,
			},
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("PublishPost", userId, parsedPostId, mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPosterService{}
			parsedPostId, _ := uuid.Parse(tt.postId)
			tt.setupMock(mockService, parsedPostId)

			controller := &PosterController{service: mockService}

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/post/%s/status", tt.postId), bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("postId", tt.postId)
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, user))

			rr := httptest.NewRecorder()
			controller.PublishHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "PublishPost")
			}
		})
	}
}

func TestPosterController_AddImageHandler(t *testing.T) {
	userId := uuid.New()
	postId := uuid.New()
	imageId := uuid.New()
	user := &dto.UserDB{UserId: userId, Role: types.Author}

	tests := []struct {
		name           string
		postId         string
		hasFile        bool
		setupMock      func(*MockPosterService, uuid.UUID)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name:    "successful add image",
			postId:  postId.String(),
			hasFile: true,
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("AddImage", userId, parsedPostId, mock.Anything, mock.Anything).
					Return(&dto.AddImageResponse{
						ImageId:  imageId,
						ImageUrl: "https://example.com/image.jpg",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.AddImageResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, imageId, resp.ImageId)
				assert.NotEmpty(t, resp.ImageUrl)
			},
		},
		{
			name:           "invalid post ID",
			postId:         "invalid-uuid",
			hasFile:        true,
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: false,
		},
		{
			name:           "no file provided",
			postId:         postId.String(),
			hasFile:        false,
			setupMock:      func(m *MockPosterService, parsedPostId uuid.UUID) {},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: false,
		},
		{
			name:    "no access",
			postId:  postId.String(),
			hasFile: true,
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("AddImage", userId, parsedPostId, mock.Anything, mock.Anything).
					Return(nil, errors.ErrorServiceNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
		},
		{
			name:    "post not found",
			postId:  postId.String(),
			hasFile: true,
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("AddImage", userId, parsedPostId, mock.Anything, mock.Anything).
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: true,
		},
		{
			name:    "incorrect data",
			postId:  postId.String(),
			hasFile: true,
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("AddImage", userId, parsedPostId, mock.Anything, mock.Anything).
					Return(nil, errors.ErrorServiceIncorrectData)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
		},
		{
			name:    "unexpected error",
			postId:  postId.String(),
			hasFile: true,
			setupMock: func(m *MockPosterService, parsedPostId uuid.UUID) {
				m.On("AddImage", userId, parsedPostId, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("storage error"))
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPosterService{}
			parsedPostId, _ := uuid.Parse(tt.postId)
			tt.setupMock(mockService, parsedPostId)

			controller := &PosterController{service: mockService}

			var req *http.Request
			if tt.hasFile {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image content"))
				writer.Close()

				req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/post/%s/images", tt.postId), body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
			} else {
				req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/post/%s/images", tt.postId), nil)
			}
			req.SetPathValue("postId", tt.postId)
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, user))

			rr := httptest.NewRecorder()
			controller.AddImageHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "AddImage")
			}
		})
	}
}

func TestPosterController_DeleteImageHandler(t *testing.T) {
	userId := uuid.New()
	postId := uuid.New()
	imageId := uuid.New()
	user := &dto.UserDB{UserId: userId, Role: types.Author}

	tests := []struct {
		name           string
		postId         string
		imageId        string
		setupMock      func(*MockPosterService, uuid.UUID, uuid.UUID)
		expectedStatus int
		checkBody      func(*testing.T, string)
		shouldCallMock bool
	}{
		{
			name:    "successful delete image",
			postId:  postId.String(),
			imageId: imageId.String(),
			setupMock: func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {
				m.On("DeleteImage", userId, parsedPostId, parsedImageId).
					Return(&dto.DeleteImageResponse{
						ImageId: parsedImageId,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldCallMock: true,
			checkBody: func(t *testing.T, body string) {
				var resp dto.DeleteImageResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, imageId, resp.ImageId)
			},
		},
		{
			name:           "invalid post ID",
			postId:         "invalid-uuid",
			imageId:        imageId.String(),
			setupMock:      func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: false,
		},
		{
			name:           "invalid image ID",
			postId:         postId.String(),
			imageId:        "invalid-uuid",
			setupMock:      func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: false,
		},
		{
			name:    "no access",
			postId:  postId.String(),
			imageId: imageId.String(),
			setupMock: func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {
				m.On("DeleteImage", userId, parsedPostId, parsedImageId).
					Return(nil, errors.ErrorServiceNoAccess)
			},
			expectedStatus: http.StatusForbidden,
			shouldCallMock: true,
		},
		{
			name:    "post not found",
			postId:  postId.String(),
			imageId: imageId.String(),
			setupMock: func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {
				m.On("DeleteImage", userId, parsedPostId, parsedImageId).
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			shouldCallMock: true,
		},
		{
			name:    "incorrect data",
			postId:  postId.String(),
			imageId: imageId.String(),
			setupMock: func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {
				m.On("DeleteImage", userId, parsedPostId, parsedImageId).
					Return(nil, errors.ErrorServiceIncorrectData)
			},
			expectedStatus: http.StatusBadRequest,
			shouldCallMock: true,
		},
		{
			name:    "unexpected error",
			postId:  postId.String(),
			imageId: imageId.String(),
			setupMock: func(m *MockPosterService, parsedPostId, parsedImageId uuid.UUID) {
				m.On("DeleteImage", userId, parsedPostId, parsedImageId).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusBadGateway,
			shouldCallMock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPosterService{}
			parsedPostId, _ := uuid.Parse(tt.postId)
			parsedImageId, _ := uuid.Parse(tt.imageId)
			tt.setupMock(mockService, parsedPostId, parsedImageId)

			controller := &PosterController{service: mockService}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/post/%s/images/%s", tt.postId, tt.imageId), nil)
			req.SetPathValue("postId", tt.postId)
			req.SetPathValue("imageId", tt.imageId)
			req = req.WithContext(context.WithValue(req.Context(), types.CtxUser, user))

			rr := httptest.NewRecorder()
			controller.DeleteImageHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Expected status %d, got %d. Response: %s",
				tt.expectedStatus, rr.Code, rr.Body.String())

			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			}

			if tt.shouldCallMock {
				mockService.AssertExpectations(t)
			} else {
				mockService.AssertNotCalled(t, "DeleteImage")
			}
		})
	}
}

func TestPosterController_EditPostHandler_NoUser(t *testing.T) {
	mockService := &MockPosterService{}
	controller := &PosterController{service: mockService}

	postId := uuid.New()
	bodyBytes, _ := json.Marshal(dto.EditPostRequest{Title: "Test", Content: "Content"})
	req := httptest.NewRequest(http.MethodPut, "/post/"+postId.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("postId", postId.String())

	rr := httptest.NewRecorder()
	controller.EditPostHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "EditPost")
}

func TestPosterController_PublishHandler_NoUser(t *testing.T) {
	mockService := &MockPosterService{}
	controller := &PosterController{service: mockService}

	postId := uuid.New()
	bodyBytes, _ := json.Marshal(dto.PublishPostRequest{Status: types.Published})
	req := httptest.NewRequest(http.MethodPatch, "/post/"+postId.String()+"/status", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("postId", postId.String())

	rr := httptest.NewRecorder()
	controller.PublishHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "PublishPost")
}

func TestPosterController_AddImageHandler_NoUser(t *testing.T) {
	mockService := &MockPosterService{}
	controller := &PosterController{service: mockService}

	postId := uuid.New()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "test.jpg")
	part.Write([]byte("fake image content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/post/"+postId.String()+"/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("postId", postId.String())

	rr := httptest.NewRecorder()
	controller.AddImageHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "AddImage")
}

func TestPosterController_DeleteImageHandler_NoUser(t *testing.T) {
	mockService := &MockPosterService{}
	controller := &PosterController{service: mockService}

	postId := uuid.New()
	imageId := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/post/"+postId.String()+"/images/"+imageId.String(), nil)
	req.SetPathValue("postId", postId.String())
	req.SetPathValue("imageId", imageId.String())

	rr := httptest.NewRecorder()
	controller.DeleteImageHandler(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), errors.ErrorHttpIncorrectUser.Error())
	mockService.AssertNotCalled(t, "DeleteImage")
}
