dev:
	@docker compose up --build
build:
	@docker build -f Dockerfile -t gin-gorm-api .
clean:
	@docker compose down --volumes
test:
	@go test -v ./...
