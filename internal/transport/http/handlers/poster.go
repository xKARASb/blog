package handlers

import (
	"database/sql"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	json "github.com/mailru/easyjson"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
	"github.com/xkarasb/blog/pkg/utils"
)

type PosterService interface {
	EditPost(userId, postId uuid.UUID, post *dto.EditPostRequest) (*dto.EditPostResponse, error)
	PublishPost(userId, postId uuid.UUID, post *dto.PublishPostRequest) (*dto.PublishPostResponse, error)
	AddImage(userId, postId uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*dto.AddImageResponse, error)
	DeleteImage(userId, postId, imageId uuid.UUID) (*dto.DeleteImageResponse, error)
}

type PosterController struct {
	service PosterService
}

func NewPosterController(service PosterService) *PosterController {
	return &PosterController{service}
}

// @Description	Add Image
// @Tags			Poster
// @Accept			mpfd
// @Produce		json
// @Security		BearerAuth
// @Param			postId	path		string	true	"Post ID"	format(uuid)
// @Param			image	formData	file	true	"Image"
// @Success		201		{object}	dto.AddImageResponse
// @Failure		400		"Incorrect body\nRefresh token expired or incorrect"
// @Failure		403		"Access denied"
// @Failure		404		"Post not found"
// @Router			/post/{postId}/images [post]“
func (c *PosterController) AddImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		return
	}
	file, fileHeader, err := r.FormFile("image")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	resp, err := c.service.AddImage(user.UserId, postId, file, fileHeader)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			http.Error(w, errors.ErrorHttpAccessDenied.Error(), http.StatusForbidden)
		case errors.ErrorServiceIncorrectData:
			http.Error(w, errors.ErrorHttpIncorrectStatus.Error(), http.StatusBadRequest)
		case sql.ErrNoRows:
			http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.MarshalToHTTPResponseWriter(resp, w)
}

// @Description	Edit post
// @Tags			Poster
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.EditPostRequest	true	"Edit post data"
// @Param			postId	path		string				true	"Post ID"	format(uuid)
// @Success		200		{object}	dto.EditPostResponse
// @Failure		400		"Incorrect body\nRefresh token expired or incorrect"
// @Failure		403		"Access denied"
// @Failure		404		"Post not found"
// @Router			/post/{postId} [put]“
func (c *PosterController) EditPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}
	reqPost := &dto.EditPostRequest{}
	if err := json.UnmarshalFromReader(r.Body, reqPost); err != nil {
		http.Error(w, errors.ErrorHttpIncorrectBody.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.Validate(reqPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		return
	}

	resPost, err := c.service.EditPost(user.UserId, postId, reqPost)
	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			http.Error(w, errors.ErrorHttpAccessDenied.Error(), http.StatusForbidden)
		case errors.ErrorServiceIncorrectData:
			http.Error(w, errors.ErrorHttpIncorrectStatus.Error(), http.StatusBadRequest)
		case sql.ErrNoRows:
			http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.MarshalToHTTPResponseWriter(resPost, w)
}

// @Description	Add Image
// @Tags			Poster
// @Produce		json
// @Security		BearerAuth
// @Param			postId	path		string	true	"Post ID"	format(uuid)
// @Param			imageId	path		string	true	"Image ID"	format(uuid)
// @Success		201		{object}	dto.DeleteImageResponse
// @Failure		400		"Incorrect body\nRefresh token expired or incorrect"
// @Failure		403		"Access denied"
// @Failure		404		"Post/Image not found"
// @Router			/post/{postId}/images/{imageId} [delete]“
func (c *PosterController) DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		return
	}
	imageId, err := uuid.Parse(r.PathValue("imageId"))

	if err != nil {
		http.Error(w, errors.ErrorHttpImageNotFound.Error(), http.StatusNotFound)
		return
	}

	resp, err := c.service.DeleteImage(user.UserId, postId, imageId)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			http.Error(w, errors.ErrorHttpAccessDenied.Error(), http.StatusForbidden)
		case errors.ErrorServiceIncorrectData:
			http.Error(w, errors.ErrorHttpIncorrectStatus.Error(), http.StatusBadRequest)
		case sql.ErrNoRows:
			http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.MarshalToHTTPResponseWriter(resp, w)
}

// @Summary		Publicate post
// @Description	Publish post
// @Tags			Poster
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.PublishPostRequest	true	"Publish post data"
// @Param			postId	path		string					true	"Post ID"	format(uuid)
// @Success		200		{object}	dto.EditPostResponse
// @Failure		400		"Incorrect body\nRefresh token expired or incorrect"
// @Failure		403		"Access denied"
// @Failure		404		"Post not found"
// @Router			/post/{postId}/status [patch]“
func (c *PosterController) PublishHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(types.CtxUser).(*dto.UserDB)
	if !ok {
		http.Error(w, errors.ErrorHttpIncorrectUser.Error(), http.StatusForbidden)
		return
	}
	reqPost := &dto.PublishPostRequest{}
	if err := json.UnmarshalFromReader(r.Body, reqPost); err != nil {
		http.Error(w, errors.ErrorHttpIncorrectBody.Error(), http.StatusBadRequest)
		return
	}

	if err := utils.Validate(reqPost); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		return
	}

	resPost, err := c.service.PublishPost(user.UserId, postId, reqPost)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			http.Error(w, errors.ErrorHttpAccessDenied.Error(), http.StatusForbidden)
		case errors.ErrorServiceIncorrectData:
			http.Error(w, errors.ErrorHttpIncorrectStatus.Error(), http.StatusBadRequest)
		case sql.ErrNoRows:
			http.Error(w, errors.ErrorHttpPostNotFound.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.MarshalToHTTPResponseWriter(resPost, w)
}
