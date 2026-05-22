FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.25.0

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/api /bin/api
COPY --from=builder /go/bin/goose /bin/goose
COPY migrations ./migrations

EXPOSE 8080

CMD ["/bin/api"]
