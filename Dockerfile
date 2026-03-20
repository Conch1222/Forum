FROM golang:1.24.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/api /app/api
COPY --from=builder /src/internal/search/mappings/posts_v1.json /app/internal/search/mappings/posts_v1.json

EXPOSE 8080

CMD ["./api"]