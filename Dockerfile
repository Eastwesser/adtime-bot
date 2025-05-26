FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache git
RUN go mod download
RUN go build -o /app/adtime-bot ./cmd/adtime/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/adtime-bot /app/adtime-bot

CMD ["/app/adtime-bot"]