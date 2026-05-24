FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -o cep-lookup ./cmd/cepfinder

FROM alpine:3.21

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /app/cep-lookup .

EXPOSE 8080

ENTRYPOINT ["./cep-lookup"]
