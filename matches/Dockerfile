FROM golang:1.13.7-alpine

WORKDIR /go/src/github.com/TempleEight/spec-golang/matches/

COPY . .

RUN apk add --no-cache git

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq

RUN go build -o matches

ENTRYPOINT ./matches

EXPOSE 80
