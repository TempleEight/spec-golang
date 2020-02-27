FROM golang:1.13.7-alpine

WORKDIR /auth

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config.json /etc/auth-service/

RUN apk add curl
RUN go build -o auth
RUN chmod +x wait-for-kong.sh

CMD ["./wait-for-kong.sh", "kong:8001", "--", "./auth"]

EXPOSE 82
