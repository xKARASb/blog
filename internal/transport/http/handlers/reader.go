package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
	"github.com/xkarasb/blog/pkg/utils"
)

type ReaderService interface {
	NewPost(authorId uuid.UUID, post *dto.CreatePostRequest) (*dto.CreatePostResponse, error)
	GetPublishedPosts() ([]*dto.GetPostResponse, error)
	GetAuthorPosts(authorId uuid.UUID) ([]*dto.GetPostResponse, error)
}

type ReaderController struct {
	service ReaderService
}

func NewReaderController(service ReaderService) *ReaderController {
	return &ReaderController{
		service: service,
	}
}

// @Summary		Read post
// @Description	Read all posts
// @Tags			Reader
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	[]dto.GetPostResponse
// @Failure		400	"Incorrect body\nRefresh token expired or incorrect"
// @Failure		403	"Access denied"
// @Failure		404	"Post not found"
// @Router			/posts [get]
func (c *ReaderController) ViewSelectionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}
	switch user.Role {
	case types.Author:
		c.authorView(w, r)
	case types.Reader:
		c.readerView(w, r)
	default:
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
	}
}

func (c *ReaderController) readerView(w http.ResponseWriter, r *http.Request) {
	posts, err := c.service.GetPublishedPosts()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (c *ReaderController) authorView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}
	posts, err := c.service.GetAuthorPosts(user.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

// @Summary		Create post
// @Description	Create new post
// @Tags			Poster
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.CreatePostRequest	true	"Create post data"
// @Success		201		{object}	dto.CreatePostResponse
// @Failure		400		"Incorrect body"
// @Failure		403		"Incorrect user"
// @Failure		409		"Idempotency key already used"
// @Router			/posts [post]
func (c *ReaderController) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}

	reqPost := &dto.CreatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(reqPost); err != nil {
		http.Error(w, errors.ErrorHttpIncorrectBody.Error(), http.StatusBadRequest)
		return

	}
	if err := utils.Validate(reqPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resPost, err := c.service.NewPost(user.UserId, reqPost)

	if err != nil {
		if err == errors.ErrorKeyIdempotencyAlreadyUsed {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resPost)

}
