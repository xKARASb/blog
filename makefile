.PHONY: swagger build json run docker-up docker-down docker-build


swagger: 
	@echo "Build swagger API"
	@swag fmt
	@swag init --parseDependency --parseInternal -g ./internal/core/servers/server.go -output ./docs  || true 


build:
	@echo "Building app..."
	@go build -o "$(CURDIR)/bin/app" "$(CURDIR)/cmd/server/main.go"

json:
	@echo "Generating dto models..."
	@easyjson -all ./internal/core/dto/

run: json swagger
	@echo "Run app.."
	@go run "$(CURDIR)/cmd/server/main.go"

docker-dev: json swagger
	docker-compose up --build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build: json swagger
	docker build -t app:latest .