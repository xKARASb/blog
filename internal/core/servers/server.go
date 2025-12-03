package servers

import (
	"fmt"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/xkarasb/blog/docs"
	"github.com/xkarasb/blog/internal/core/repository"
	"github.com/xkarasb/blog/internal/core/service"
	"github.com/xkarasb/blog/internal/transport/http/middlewares"
	"github.com/xkarasb/blog/internal/transport/http/routers"
	"github.com/xkarasb/blog/pkg/db/postgres"
	// Your module name + /docs
)

type HttpServerConfig struct {
	Address string `env:"ADDRESS" env-default:"127.0.0.1"`
	Port    int    `env:"PORT" env-default:"8080"`
}

type HttpServer struct {
	cfg  *HttpServerConfig
	http *http.Server
}

func NewHttpServer(cfg *HttpServerConfig, db *postgres.DB, isDoc bool) *HttpServer {
	apiRouter := http.NewServeMux()

	dbRepo := repository.NewBlogRepository(db)

	authService := service.NewAuthService(dbRepo, "secret")
	readerService := service.NewReaderService(dbRepo)
	posterService := service.NewPosterService(dbRepo)

	authMMan := middlewares.NewAuthMiddlewareManager(authService) //AuthMiddleWareManager - создаёт объект, где хранится секрет, для более гибкой работы с мидлварами и передачи их в роутеры

	authRouter := routers.GetAuthRouter(authService)
	readRouter := routers.GetReaderRouter(readerService, authMMan)
	posterRouter := routers.GetPosterRouter(posterService)

	apiRouter.Handle("/", authMMan.AuthMiddleware(readRouter))
	// Поменял ендпоинт т.к стандартный пакет не может сравнивать схожие ендпоинты в разных роутерах, что приводит к неверному поведению
	apiRouter.Handle("/post/", authMMan.AuthMiddleware(authMMan.AuthorOnlyMiddleware(posterRouter)))
	apiRouter.Handle("/auth/", authRouter)

	http.DefaultServeMux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", cfg.Address, cfg.Port),
	}

	if isDoc {
		docs.SwaggerInfo.Title = "CPC Blog API"
		docs.SwaggerInfo.Description = "This is API CPC Blog server"
		docs.SwaggerInfo.Version = "1.0"
		docs.SwaggerInfo.Host = server.Addr
		docs.SwaggerInfo.BasePath = "/api"

		http.Handle("/swagger/", httpSwagger.WrapHandler)
	}
	fmt.Println(server.Addr)

	return &HttpServer{
		cfg,
		server,
	}
}

func (s *HttpServer) Start() error {
	return s.http.ListenAndServe()
}

func (s *HttpServer) Stop() error {
	return s.http.Close()
}
