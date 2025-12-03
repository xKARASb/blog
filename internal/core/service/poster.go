package service

import (
	"github.com/google/uuid"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
)

type PosterRepository interface {
	GetPostByIdempotencyKey(idempotencyKey string) (*dto.PostDB, error)
	GetPostById(id uuid.UUID) (*dto.PostDB, error)
	UpdatePost(id uuid.UUID, title, content string, status types.PostStatus) (*dto.PostDB, error)
}

type PosterService struct {
	rep PosterRepository
}

func NewPosterService(rep PosterRepository) *PosterService {
	return &PosterService{rep}
}

func (s *PosterService) EditPost(userId, postId uuid.UUID, post *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	postDB, err := s.rep.GetPostById(postId)

	if err != nil {
		return nil, err
	}
	if postDB.AuthorId != userId {
		return nil, errors.ErrorServiceNoAccess
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
	postDB, err := s.rep.GetPostById(postId)

	if err != nil {
		return nil, err
	}
	if postDB.AuthorId != userId {
		return nil, errors.ErrorServiceNoAccess
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
