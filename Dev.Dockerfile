FROM golang:1.23-alpine

RUN mkdir -p /run/secrets

WORKDIR /src

RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download
RUN go get github.com/swaggo/swag

CMD ["air", "-c", ".air.toml"]
