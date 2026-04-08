# Dockerfile for IFMS-BE-API
# Build stage
FROM golang:1.26.1-alpine AS builder

run apk add --no-cache git

ENV GOPRIVATE=github.com/IFMS-Team/*

WORKDIR /app

# Approve Go download repo from Organization
ARG GITHUB_TOKEN
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api/main.go

# Run stage

FROM alpine:latest AS runner

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./server"]