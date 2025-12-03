package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/xkarasb/blog/pkg/types"
)

// @Description	Post database model with all fields
//
//easyjson:skip
type PostDB struct {
	PostId         uuid.UUID        `json:"post_id" db:"post_id"`
	AuthorId       uuid.UUID        `json:"author_id" db:"author_id"`
	IdempotencyKey string           `json:"indempotency_key" db:"idempotency_key"`
	Title          string           `json:"title" db:"title"`
	Content        string           `json:"content" db:"content"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
	Status         types.PostStatus `json:"status" db:"status"`
} //	@name	Post

// @Description	Request payload for creating a new post
type CreatePostRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Title          string `json:"title"`
	Content        string `json:"content"`
} //	@name	CreatePostRequest

// @Description	Response with ID of the created post
type CreatePostResponse struct {
	PostId uuid.UUID `json:"post_id"`
} //	@name	CreatePostResponse

// @Description	Request payload for editing a post
type EditPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
} //	@name	EditPostRequest

// @Description	Response with updated post details
type EditPostResponse struct {
	PostId         uuid.UUID        `json:"post_id"`
	AuthorId       uuid.UUID        `json:"author_id"`
	IdempotencyKey string           `json:"indempotency_key"`
	Title          string           `json:"title"`
	Content        string           `json:"content"`
	Status         types.PostStatus `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
} //	@name	PostDetails

// @Description	Request to change post status (publish/unpublish)
type PublishPostRequest struct {
	Status types.PostStatus `json:"status"`
} //	@name	UpdatePostStatusRequest

// @Description	Response with ID of the published post
type PublishPostResponse struct {
	PostId uuid.UUID `json:"post_id"`
} //	@name	UpdatePostStatusResponse
