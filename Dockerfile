FROM golang:1.18.4-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -v -o server /app/cmd/main.go

CMD ["/app/server" ]