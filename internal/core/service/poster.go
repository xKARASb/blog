package service

import (
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
)

type PosterRepository interface {
	GetPostByIdempotencyKey(idempotencyKey string) (*dto.PostDB, error)
	GetPostById(id uuid.UUID) (*dto.PostDB, error)
	UpdatePost(id uuid.UUID, title, content string, status types.PostStatus) (*dto.PostDB, error)
	CreateImage(imageId, postId uuid.UUID, imageUrl string) (*dto.ImageDB, error)
	DeleteImage(imageId uuid.UUID) (*dto.ImageDB, error)
}

type PosterStorageRepositry interface {
	PutImage(fileName string, file io.Reader, fileSize int64, contentType string) (string, error)
	DeleteImage(objectName string) error
}

type PosterService struct {
	rep  PosterRepository
	stor PosterStorageRepositry
}

func NewPosterService(rep PosterRepository, stor PosterStorageRepositry) *PosterService {
	return &PosterService{rep, stor}
}

func (s *PosterService) getPostAuthor(userId, postId uuid.UUID) (*dto.PostDB, error) {
	postDB, err := s.rep.GetPostById(postId)

	if err != nil {
		return nil, err
	}
	if postDB.AuthorId != userId {
		return nil, errors.ErrorServiceNoAccess
	}
	return postDB, nil
}

func (s *PosterService) EditPost(userId, postId uuid.UUID, post *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	postDB, err := s.getPostAuthor(userId, postId)

	if err != nil {
		return nil, err
	}

	postDB, err = s.rep.UpdatePost(postId, post.Title, post.Content, postDB.Status)
	if err != nil {
		return nil, err
	}

	postRes := &dto.EditPostResponse{
		PostId:         postDB.PostId,
		AuthorId:       postDB.AuthorId,
		IdempotencyKey: postDB.IdempotencyKey,
		Title:          postDB.Title,
		Content:        postDB.Content,
		Status:         postDB.Status,
		CreatedAt:      postDB.CreatedAt,
		UpdatedAt:      postDB.UpdatedAt,
	}
	return postRes, nil
}
func (s *PosterService) PublishPost(userId, postId uuid.UUID, post *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	postDB, err := s.getPostAuthor(userId, postId)

	if err != nil {
		return nil, err
	}

	if post.Status != types.Published {
		return nil, errors.ErrorServiceIncorrectData
	}

	postDB, err = s.rep.UpdatePost(postId, postDB.Title, postDB.Content, post.Status)
	if err != nil {
		return nil, err
	}

	postRes := &dto.PublishPostResponse{
		PostId: postDB.PostId,
	}
	return postRes, nil
}

func (s *PosterService) AddImage(userId, postId uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*dto.AddImageResponse, error) {
	_, err := s.getPostAuthor(userId, postId)

	if err != nil {
		return nil, err
	}

	size := fileHeader.Size
	contentType := fileHeader.Header.Get("Content-Type")

	imageId, err := uuid.NewUUID()

	if err != nil {
		return nil, err
	}

	link, err := s.stor.PutImage(imageId.String(), file, size, contentType)
	if err != nil {
		return nil, err
	}
	imageDB, err := s.rep.CreateImage(imageId, postId, link)

	if err != nil {
		return nil, err
	}

	imageRes := &dto.AddImageResponse{
		ImageId:  imageDB.ImageId,
		ImageUrl: link,
	}

	return imageRes, nil
}

func (s *PosterService) DeleteImage(userId, postId, imageId uuid.UUID) (*dto.DeleteImageResponse, error) {
	_, err := s.getPostAuthor(userId, postId)

	if err != nil {
		return nil, err
	}

	if _, err = s.rep.DeleteImage(imageId); err != nil {
		return nil, err
	}

	if err = s.stor.DeleteImage(imageId.String()); err != nil {
		return nil, err
	}

	return &dto.DeleteImageResponse{ImageId: imageId}, nil

}
