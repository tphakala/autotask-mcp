# autotask-mcp

A [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server for [Kaseya Autotask PSA](https://www.autotask.com/), written in Go.

Provides AI assistants (Claude, GPT, etc.) with structured access to Autotask data and operations — tickets, companies, contacts, projects, time entries, billing, quotes, and more.

## Features

- **59 MCP tools** covering all major Autotask entity types
- **4 lazy loading meta-tools** for progressive tool discovery
- **7 MCP resource templates** for direct data access
- **Dual transport**: stdio (Claude Desktop/Code) + HTTP (gateway deployments)
- **Rate limiting**: token bucket (5000 req/hour) + 3-thread concurrency limiter
- **Circuit breaker**: automatic backoff on API failures
- **ID-to-name mapping cache** with batch preloading and 30-minute per-entry TTL
- **Compact response formatting** to minimize LLM context usage

## Quick Start

### Stdio (Claude Desktop / Claude Code)

```bash
# Set credentials
export AUTOTASK_USERNAME=api_user@company.com
export AUTOTASK_SECRET=your_secret
export AUTOTASK_INTEGRATION_CODE=YOUR_CODE

# Run
go run .
```

### Claude Desktop Configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "autotask": {
      "command": "/path/to/autotask-mcp",
      "env": {
        "AUTOTASK_USERNAME": "api_user@company.com",
        "AUTOTASK_SECRET": "your_secret",
        "AUTOTASK_INTEGRATION_CODE": "YOUR_CODE"
      }
    }
  }
}
```

### HTTP / Container

```bash
cp .env.example .env
# Edit .env with your credentials

# With Podman
podman-compose up -d

# With Docker
docker compose up -d
```

The server starts on port 8080 with endpoints:
- `POST /mcp` — MCP Streamable HTTP transport
- `GET /health` — Health check

### Gateway Mode

For multi-tenant deployments, set `AUTH_MODE=gateway`. Credentials are injected per-request via headers:

```
X-API-Key: username
X-API-Secret: secret
X-Integration-Code: code
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTOTASK_USERNAME` | — | Autotask API username (required for env mode) |
| `AUTOTASK_SECRET` | — | Autotask API secret (required for env mode) |
| `AUTOTASK_INTEGRATION_CODE` | — | Autotask integration code (required for env mode) |
| `AUTOTASK_API_URL` | auto-discovered | Override API base URL |
| `MCP_TRANSPORT` | `stdio` | Transport: `stdio` or `http` |
| `MCP_HTTP_PORT` | `8080` | HTTP server port |
| `MCP_HTTP_HOST` | `0.0.0.0` | HTTP server bind address |
| `AUTH_MODE` | `env` | Authentication: `env` or `gateway` |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `LAZY_LOADING` | `false` | Enable progressive tool discovery |

## Available Tools

### Core Entities
| Category | Tools |
|----------|-------|
| Tickets | search, get details, create, update |
| Companies | search, create, update |
| Contacts | search, create |
| Projects | search, create |
| Tasks | search, create |
| Resources | search |
| Time Entries | create, search |
| Configuration Items | search |
| Contracts | search |

### Notes & Attachments
| Category | Tools |
|----------|-------|
| Ticket Notes | get, search, create |
| Project Notes | get, search, create |
| Company Notes | get, search, create |
| Ticket Attachments | get, search |

### Financial
| Category | Tools |
|----------|-------|
| Quotes | get, search, create |
| Quote Items | get, search, create, update, delete |
| Opportunities | get, search, create |
| Invoices | search |
| Billing Items | get, search |
| Billing Approvals | search |
| Expense Reports | get, search, create |
| Expense Items | create |

### Catalog
| Category | Tools |
|----------|-------|
| Products | get, search |
| Services | get, search |
| Service Bundles | get, search |

### Utility
| Category | Tools |
|----------|-------|
| Connection | test |
| Picklists | queues, statuses, priorities, field info |
| Meta-tools | list categories, list tools, execute, router |

## Architecture

```
autotask-mcp
├── main.go / config.go / server.go    # Entry point, config, transport setup
├── services/
│   ├── mapping.go                      # Company/resource name cache (batch preload)
│   ├── picklist.go                     # Lazy-loaded field/picklist cache
│   └── formatter.go                    # Compact response formatting
├── tools/                              # 59 MCP tool handlers
│   ├── tickets.go, companies.go, ...   # Entity CRUD tools
│   ├── register.go                     # Shared helpers
│   └── lazy.go                         # Progressive discovery meta-tools
└── resources/                          # 7 MCP resource templates
```

Built on:
- [go-autotask](https://github.com/tphakala/go-autotask) — Autotask REST API client with typed generics
- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) — Official MCP server framework

## Building

```bash
go build -o autotask-mcp .
```

## Development

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Vet
go vet ./...
```

## License

Apache 2.0
