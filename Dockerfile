# 1. Fase de Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN GOOS="linux" CGO_ENABLED=0 go build -ldflags="-w -s" -o server ./cmd/server/main.go

FROM scratch
COPY --from=builder /app/server .
COPY --from=builder /app/example.env ./example.env

EXPOSE 8000 8080 50051 8082

CMD ["./server"]