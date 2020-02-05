FROM golang:1.13.7-alpine

WORKDIR /go/src/github.com/TempleEight/spec-golang/user/

COPY . .

RUN apk add --no-cache git

RUN go get github.com/gorilla/mux

RUN go build -o user

ENTRYPOINT ./user

EXPOSE 80
