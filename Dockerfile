FROM golang:1.24.4-alpine3.22 AS builder

WORKDIR /app

COPY . .

RUN go build -o app


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]
