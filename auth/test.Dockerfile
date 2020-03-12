FROM golang:1.13.7-alpine

WORKDIR /auth

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config.json /etc/auth-service/

ENTRYPOINT CGO_ENABLED=0 go test -v -tags=it