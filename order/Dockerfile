FROM golang:1.24.3 AS builder

COPY ./shared /shared

WORKDIR /app

COPY order/go.mod order/go.sum ./

RUN go mod download

COPY ./order .

RUN go build -o main cmd/order/main.go

FROM debian:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 5000

CMD ["./main"]
