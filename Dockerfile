FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o autotask-mcp .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/autotask-mcp /autotask-mcp
RUN adduser -D -u 10001 appuser
USER appuser
ENV MCP_TRANSPORT=http
ENV MCP_HTTP_PORT=8080
ENV MCP_HTTP_HOST=0.0.0.0
EXPOSE 8080
ENTRYPOINT ["/autotask-mcp"]
