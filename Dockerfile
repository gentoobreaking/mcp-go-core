# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build production binary (CGO disabled for static linking)
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=v0.1.0" \
    -o /bin/mcp-go-core \
    ./cmd/mcp-go-core

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/mcp-go-core /usr/local/bin/mcp-go-core

# Default: stdio transport (for MCP client integration)
ENTRYPOINT ["mcp-go-core"]
CMD ["run", "--transport", "stdio"]

LABEL org.opencontainers.image.title="mcp-go-core"
LABEL org.opencontainers.image.description="Modular MCP server framework"
LABEL org.opencontainers.image.source="https://github.com/project/mcp-go-core"
