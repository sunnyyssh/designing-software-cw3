FROM golang:1.24.3 AS builder

WORKDIR /app

COPY gateway/go.mod gateway/go.sum ./

RUN go mod download

COPY ./gateway .

RUN go build -o main cmd/gateway/main.go

FROM debian:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 80

CMD ["./main"]
