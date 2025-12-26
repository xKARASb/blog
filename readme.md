
# CPC Blog API

A minimal, high-performance Go backend for a blog system. This project features JWT authentication, role-based access control (Author/Reader), idempotent post creation, and image management using MinIO.

## 🚀 Features

- **Authentication**: Register, Login, and Refresh Token using JWT.
- **Role-Based Access**: 
    - **Author**: Create, edit, publish, and manage images for their posts.
    - **Reader**: Browse published blog posts.
- **Media Management**: Upload and delete post images via MinIO (S3 compatible).
- **Reliability**: Uses idempotency keys for post creation to prevent duplicate entries.
- **Performance**: Utilizes `easyjson` for optimized JSON (un)marshaling.
- **Documentation**: Fully documented with Swagger (OpenAPI 2.0).


## 🛠 Tech Stack

- **Language**: Go
- **Database**: PostgreSQL
- **Object Storage**: MinIO
- **Documentation**: Swagger (swag)
- **JSON Optimization**: EasyJSON
- **Containerization**: Docker & Docker Compose

---

## 📋 Prerequisites

Ensure you have the following installed:
- [Go](https://go.dev/) (1.21+)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [swag](https://github.com/swaggo/swag) (for API docs)
- [easyjson](https://github.com/mailru/easyjson) (for DTO generation)

---

## ⚙️ Configuration

1. Clone the [repository](https://github.com/xKARASb/blog).
    ```bash
    git clone https://github.com/xKARASb/blog && cd blog
    ```
2. Create a `.env` and `docker.env` file based on the provided example:
   ```bash
   cp example.env .env
   cp example.env docker.env
   ```
3. Update the `.env` and `docker.env` values (especially the `POSTGRES_HOST` and `MINIO_ENDPOINT` if running locally vs. in Docker).

---

## 🏃 Running the Application

### Using Docker (Recommended)
This is the fastest way to get the database, storage, and app running.
```bash
# Install dependencies
make utils
# Build and start all services
make docker-dev
# Build and start all services in background
make docker-up
```

### Local Development
If you want to run the Go server locally:
1. Make sure that the `database` and `storage` are up
2. Set up `.env` like in `example.env`
3. Install dependencies:
    ```bash
    make utils
    go mod download
    ```
4. Run the app: `make run`

---

## 📖 API Documentation

Once the server is running, you can access the interactive Swagger UI:
`http://localhost/swagger/` (Base path: `/api`)

Or with defined port if run localy 
`http://localhost:8080/swagger/` (Base path: `/api`)

To manually regenerate documentation after changing code:
```bash
make swagger
```

---

## 🛠 Available Commands

| Command | Description |
| :--- | :--- |
| `make swagger` | Update Swagger/OpenAPI documentation |
| `make json` | Generate optimized DTO models using `easyjson` |
| `make build` | Compile the binary to `./bin/app` |
| `make run` | Generate dependencies and run the server |
| `make test` | Run unit tests for HTTP handlers |
| `make docker-dev` | Build and start everything in development mode |
| `make docker-up` | Start infrastructure (DB, MinIO) in background |
| `make docker-down` | Stop and remove all containers |
| `make docker-build` | Build the production-ready Docker image |

---

## 📂 Project Structure (Partial)

- `cmd/server/main.go`: Application entry point.
- `internal/core/servers/`: Server initialization and middleware.
- `internal/core/dto/`: Data Transfer Objects (optimized with EasyJSON).
- `internal/transport/http/handlers/`: API route handlers and logic.
- `docs/`: Generated Swagger documentation.

---

## 🔒 Security

- **JWT**: Headers should include `Authorization: Bearer <your_token>`.
- **Validation**: Strict input validation on registration and post creation.
- **Storage**: MinIO access keys managed via environment variables.

---

## 📝 Common Workflows

- **Adding a new DTO**: Add the struct to `internal/core/dto/` and run `make json`.
- **Adding a new Route**: Add the logic in handlers, update Swagger comments, and run `make swagger`.
- **Deployment**: Run `make docker-build` to ensure tests pass and images are ready for registry.

---