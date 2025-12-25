package main

import (
	"log/slog"

	"github.com/xkarasb/blog/internal/config"
	"github.com/xkarasb/blog/internal/core/servers"
	"github.com/xkarasb/blog/pkg/db/postgres"
	"github.com/xkarasb/blog/pkg/storage/minio"
)

func main() {
	appCfg, err := config.NewConfig()
	db, err := postgres.New(appCfg.PostgresConfig)
	storage, err := minio.NewMinIOClient(appCfg.MinIOConfig)

	if err != nil {
		panic(err)
	}

	serv := servers.NewHttpServer(appCfg.HttpServerConfig, db, storage, appCfg.Docs)

	if err = serv.Start(); err != nil {
		slog.Error(err.Error())
	}
}
