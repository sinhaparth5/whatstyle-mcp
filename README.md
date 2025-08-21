# WhatsStyle MCP Server

A Model Context Protocol (MCP) server with Grok AI integration for intelligent WhatsApp messaging.

## Features

- **MCP Protocol** - Official Go SDK compliance
- **Grok AI** - X.AI powered responses with conversation context
- **WhatsApp Ready** - Business API webhook support
- **SQLite Database** - Message persistence and user management
- **Production Ready** - Docker, CI/CD, monitoring

## Quick Start

### Prerequisites
- Go 1.21+
- SQLite3
- [Grok API key](https://x.ai) from X.AI

### Installation
```bash
git clone https://github.com/sinhaparth5/whatstyle-mcp.git
cd whatstyle-mcp
go mod download
```

### Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit with your API keys
GROK_API_KEY=xai-your_actual_key_here
PORT=8080
```

### Run
```bash
# Build and run
make build
./bin/mcp-server

# Or run directly
go run cmd/server/main.go
```

## API Endpoints

- `GET /health` - Server health check
- `POST /mcp` - MCP protocol endpoint
- `GET /tools` - Available MCP tools
- `POST /webhook` - WhatsApp webhook handler

## MCP Tools

### Chat Tool
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "chat",
    "arguments": {
      "user_id": "user123",
      "message": "Hello!"
    }
  }
}
```

### History Tool
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "history",
    "arguments": {
      "user_id": "user123",
      "limit": 10
    }
  }
}
```

## Development

### Quality Checks
```bash
make quality      # All checks + 80% coverage
make lint         # Code linting
make test         # Run tests
make security     # Security scan
```

### Coverage
```bash
make test-coverage           # Generate HTML report
make test-coverage-threshold # Check 80% minimum
```

## Deployment

### Docker
```bash
docker build -t whatsapp-mcp-server .
docker run -p 8080:8080 -e GROK_API_KEY=your_key whatsapp-mcp-server
```

### Railway/Heroku
Set environment variables and deploy from GitHub.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GROK_API_KEY` | X.AI Grok API key | Required |
| `PORT` | Server port | 8080 |
| `DATABASE_PATH` | SQLite database path | `./mcp_server.db` |
| `GROK_MODEL` | Grok model to use | `grok-beta` |

## Architecture

```
WhatsApp → Webhook → MCP Server → Grok AI → Response
                         ↓
                   SQLite Database
```

## License

Licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

## Support

- GitHub Issues for bugs and features
- [X.AI Documentation](https://docs.x.ai) for Grok API
- [MCP Specification](https://spec.modelcontextprotocol.io) for protocol details