FROM golang:1.19.1 as builder
ENV GO111MODULE=on
WORKDIR /src
COPY . .

RUN go mod download

RUN go build -v -ldflags '-extldflags "-static"' -o server /src/cmd/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=builder /src/server /app/server

RUN chmod +x /app/server

ENTRYPOINT ["/app/server"]