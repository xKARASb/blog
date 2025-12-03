package handlers

import (
	"fmt"
	"net/http"

	json "github.com/mailru/easyjson"

	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/errors"
)

type AuthService interface {
	RegistrateUser(user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error)
	LoginUser(user *dto.LoginUserRequest) (*dto.LoginUserResponse, error)
	RefreshToken(token *dto.RefreshRequest) (*dto.RefreshResponse, error)
	AuthorizeUser(token string) (*dto.UserDB, error)
}

type AuthController struct {
	service AuthService
}

func NewAuthController(service AuthService) *AuthController {
	return &AuthController{service: service}
}

// @Summary		Registration
// @Description	Registrate a new user
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RegistrateUserRequest	true	"Registration data"
// @Success		200		{object}	dto.RegistrateUserResponse
// @Failure		403		"User alredy exsist"
// @Failure		400		"Incorrect email format\nIncorrect body"
// @Router			/auth/register [post]
func (c *AuthController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &dto.RegistrateUserRequest{}

	if err := json.UnmarshalFromReader(r.Body, reqUser); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect body")
		return
	}
	resp, err := c.service.RegistrateUser(reqUser)
	if err != nil {
		if err == errors.ErrorRepositoryUserAlreadyExsist {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "%s\n", err.Error())
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.MarshalToHTTPResponseWriter(resp, w)
}

// @Summary		Login
// @Description	Login a user
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.LoginUserRequest	true	"Login data"
// @Success		200		{object}	dto.LoginUserResponse
// @Failure		400		"Incorrect body"
// @Failure		403		"Email or password incorrect"
// @Router			/auth/login [post]
func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	reqUser := &dto.LoginUserRequest{}
	if err := json.UnmarshalFromReader(r.Body, reqUser); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect body")
		return
	}
	resp, err := c.service.LoginUser(reqUser)

	if err != nil {
		if err == errors.ErrorRepositoryEmailNotExsist {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Email or password incorrect\n")
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.MarshalToHTTPResponseWriter(resp, w)
}

// @Summary		Invoke refresh token
// @Description	Get access token by refresh token
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RefreshRequest	true	"Refresh token data"
// @Success		200		{object}	dto.RefreshResponse
// @Failure		400		"Incorrect body\nRefresh token expired or incorrect"
// @Router			/auth/refresh-token [post]
func (c *AuthController) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.RefreshRequest{}
	if err := json.UnmarshalFromReader(r.Body, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Incorrect body")
		return
	}
	resp, err := c.service.RefreshToken(req)

	if err != nil {
		if err == errors.ErrorInvalidToken {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Refresh token expired or incorrect\n")
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.MarshalToHTTPResponseWriter(resp, w)
}
