FROM golang:1.13.7-alpine

WORKDIR /match

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config.json /etc/match-service/

RUN go build -o match

ENTRYPOINT ./match

EXPOSE 81
