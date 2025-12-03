package dto

import (
	"time"

	"github.com/google/uuid"
)

//easyjson:skip
type ImageDB struct {
	ImageID   uuid.UUID `json:"image_id" db:"image_id"`
	PostId    uuid.UUID `json:"post_id" db:"post_id"`
	ImageUrl  string    `json:"image_url" db:"image_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
