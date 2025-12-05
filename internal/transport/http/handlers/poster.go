package handlers

import (
	"database/sql"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	json "github.com/mailru/easyjson"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Incorrect user")
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Post not found")
		return
	}
	file, fileHeader, err := r.FormFile("image")

	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, "Some error while get file")
		return
	}

	resp, err := c.service.AddImage(user.UserId, postId, file, fileHeader)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Access denied")
		case errors.ErrorServiceIncorrectData:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect status")
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Post not found")
		default:
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "Something wrong")
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Incorrect user")
		return
	}
	reqPost := &dto.EditPostRequest{}
	if err := json.UnmarshalFromReader(r.Body, reqPost); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect body")
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Post not found")
		return
	}

	resPost, err := c.service.EditPost(user.UserId, postId, reqPost)
	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Access denied")
		case errors.ErrorServiceIncorrectData:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect status")
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Post not found")
		default:
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "Something wrong")
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Incorrect user")
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Post not found")
		return
	}
	imageId, err := uuid.Parse(r.PathValue("imageId"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Image not found")
		return
	}

	resp, err := c.service.DeleteImage(user.UserId, postId, imageId)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Access denied")
		case errors.ErrorServiceIncorrectData:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect status")
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Post not found")
		default:
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "Something wrong")
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Incorrect user")
		return
	}
	reqPost := &dto.PublishPostRequest{}
	if err := json.UnmarshalFromReader(r.Body, reqPost); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect body")
		return
	}

	postId, err := uuid.Parse(r.PathValue("postId"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Post not exsist")
		return
	}

	resPost, err := c.service.PublishPost(user.UserId, postId, reqPost)

	if err != nil {
		switch err {
		case errors.ErrorServiceNoAccess:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Access denied")
		case errors.ErrorServiceIncorrectData:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Incorrect status")
		case sql.ErrNoRows:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Post not found")
		default:
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "Something wrong")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.MarshalToHTTPResponseWriter(resPost, w)
}
