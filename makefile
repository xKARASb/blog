.PHONY: swagger
swagger: 
	@echo "Build swagger API"
	@swag fmt
	@swag init --parseDependency --parseInternal -g ./internal/core/servers/server.go -output ./docs  || true 


.PHONY: build
build:
	@echo "Building app..."
	@go build -o "$(CURDIR)/bin/app" "$(CURDIR)/cmd/server/main.go"

.PHONY: dto
dto:
	@echo "Generating dto models..."
	@easyjson -all ./internal/core/dto/

.PHONY: run
run: dto swagger
	@echo "Run app.."
	@go run "$(CURDIR)/cmd/server/main.go"