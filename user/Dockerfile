FROM golang:1.13.7-alpine

WORKDIR /user

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config.json /etc/user-service/

RUN go build -o user 

ENTRYPOINT ./user

EXPOSE 80

