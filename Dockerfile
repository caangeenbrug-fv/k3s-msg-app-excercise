FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine3.22 AS builder
WORKDIR /src
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app ./cmd/server

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /src/app .
EXPOSE 8080
CMD ["./app"]
