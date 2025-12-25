.PHONY: swagger build json run docker-up docker-down docker-build utils test help

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

utils:
	@echo "Installing deps..."
	@echo "swag installing..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "easyjson installing..."
	@go install github.com/mailru/easyjson/...@latest

test:
	@go test ./internal/transport/http/handlers/... -v

docker-dev: json swagger
	docker-compose up --build

docker-up: test json swagger
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-build: test json swagger
	docker build -t app:latest .

help:
	@echo "Available commands:"
	@echo "  make swagger      - Generate/update Swagger/OpenAPI documentation"
	@echo "  make build        - Build the application binary to ./bin/app"
	@echo "  make json         - Generate DTO models using easyjson"
	@echo "  make run          - Run the application (generates JSON and Swagger first)"
	@echo "  make test         - Run HTTP handler tests"
	@echo "  make docker-dev   - Build and start containers in development mode"
	@echo "  make docker-up    - Start containers in detached mode"
	@echo "  make docker-down  - Stop and remove containers"
	@echo "  make docker-build - Build Docker image (runs tests, JSON gen, Swagger first)"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Common workflows:"
	@echo "  Development:        make run"
	@echo "  Build & test:       make build && make test"
	@echo "  Docker development: make docker-dev"
	@echo "  Full deployment:    make docker-build"
	@echo "Don't forget correct setup docker.env for container running"