# Build stage
FROM golang:1.26.1-trixie AS builder

ARG VERSION=latest

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -ldflags "-X main.FullVersion=${VERSION}" -o model-router .

# Runtime stage
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/model-router .

USER nonroot

EXPOSE 12345

ENTRYPOINT ["./model-router"]
