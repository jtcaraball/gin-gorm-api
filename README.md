# Gin & Gorm API
An API boilerplate using Gin, Gorm and Docker.

## Features
- Token based authentication scheme using HMAC-SHA256.
- Custom scheme validation using middleware.
- Live reloading.
- PostgreSQL database for development, available through docker compose.
- Swagger UI available at http://localhost:8080/swagger/index.html.

## Dependencies
- [gin](https://github.com/gin-gonic/gin) as web framework.
- [gorm](https://github.com/go-gorm/gorm) as ORM.
- [ozzo-validation](https://github.com/go-ozzo/ozzo-validation) for form
  validation.
- [gin-swagger](https://github.com/swaggo/gin-swagger) for OpenAPI spec
  generation.
- [go-envconfig](https://github.com/sethvargo/go-envconfig) and
  [yaml](https://github.com/go-yaml/yaml/tree/v3) for configuration management.
- [air](https://github.com/air-verse/air) for live reloading.
- [golancilint](https://github.com/golangci/golangci-lint) for linting.

## Commands
- `make dev`: To run the project locally.
- `make build`: To build the project as a docker image.
- `make clean`: To remove the database volume.
