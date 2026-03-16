FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/todo-server ./cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app-bin .
COPY --from=builder /app/web ./web
RUN chmod +x ./todo-server
COPY .env . 
EXPOSE 8080
CMD ["./todo-server"]