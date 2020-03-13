FROM golang:1.13.7-alpine

WORKDIR /match

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config.json /etc/match-service/

ENTRYPOINT CGO_ENABLED=0 go test -v -tags=it
