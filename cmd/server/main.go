package main

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/xkarasb/blog/internal/core/servers"
	"github.com/xkarasb/blog/pkg/db/postgres"
	"github.com/xkarasb/blog/pkg/storage/minio"
)

func main() {
	httpCfg := servers.HttpServerConfig{}
	dbCfg := postgres.PostgresConfig{}
	storageCfg := minio.MinIOConfig{}
	cleanenv.ReadConfig(".env", &httpCfg)
	cleanenv.ReadConfig(".env", &dbCfg)
	cleanenv.ReadConfig(".env", &storageCfg)
	db, err := postgres.New(&dbCfg)
	storage, err := minio.NewMinIOClient(storageCfg)

	if err != nil {
		panic(err)
	}

	serv := servers.NewHttpServer(&httpCfg, db, storage, true)

	fmt.Println(serv.Start())
}
