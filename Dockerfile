FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/attendance-bot ./cmd/bot

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/attendance-bot /app/attendance-bot
COPY credentials.json /app/credentials.json

ENV GOOGLE_CREDENTIALS_PATH=/app/credentials.json

ENTRYPOINT ["/app/attendance-bot"]
