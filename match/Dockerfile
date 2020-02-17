FROM golang:1.13.7-alpine

WORKDIR /go/src/github.com/TempleEight/spec-golang/match/

COPY . .
COPY config.json /etc/match-service/

RUN apk add --no-cache git

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq
RUN go get github.com/asaskevich/govalidator

RUN go build -o match

ENTRYPOINT ./match

EXPOSE 81
