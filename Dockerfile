# Build stage
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-X main.FullVersion=latest" -o model-router .

# Runtime stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/model-router .

EXPOSE 12345

ENTRYPOINT ["./model-router"]
