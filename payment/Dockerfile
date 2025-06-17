FROM golang:1.24.3 AS builder

COPY ./shared /shared

WORKDIR /app

COPY gateway/go.mod gateway/go.sum ./

RUN go mod download

COPY ./payment .

RUN go build -o main cmd/payment/main.go

FROM debian:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 5000

CMD ["./main"]
