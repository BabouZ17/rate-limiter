FROM golang:1.18.4-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY config/redis_config.json /app/config/config.json

RUN go mod download

COPY . ./

RUN go build -v -o server /app/cmd/main.go

CMD ["/app/server" ]