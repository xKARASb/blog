package routers

import (
	"net/http"

	"github.com/xkarasb/blog/internal/core/service"
	"github.com/xkarasb/blog/internal/transport/http/handlers"
)

func GetPosterRouter(service *service.PosterService) *http.ServeMux {
	controller := handlers.NewPosterController(service)
	router := http.NewServeMux()

	router.HandleFunc("POST /post/{postId}/images", controller.AddImageHandler)
	router.HandleFunc("PUT /post/{postId}", controller.EditPostHandler)
	router.HandleFunc("DELETE /post/{postId}/images/{imageId}", controller.DeleteImageHandler)
	router.HandleFunc("PATCH /post/{postId}/status", controller.PublishHandler)

	return router
}
