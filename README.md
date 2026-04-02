# agent-probe

[![CI](https://github.com/arcane-bear/agent-probe/actions/workflows/ci.yml/badge.svg)](https://github.com/arcane-bear/agent-probe/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/arcane-bear/agent-probe)](https://goreportcard.com/report/github.com/arcane-bear/agent-probe)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**A lightweight Go tool for probing and monitoring AI agent health and performance.** Verifies LLM connectivity, token budgets, tool availability, and response quality — beyond simple HTTP 200.

Built by [RapidClaw](https://rapidclaw.dev) — self-hosted AI agent deployment platform.

## The Problem

Your Kubernetes liveness probe returns `200 OK`. Your agent is "healthy." But:

- Your LLM API key expired 3 hours ago
- You've burned through 98% of your token budget
- The vector database your agent depends on is unreachable
- Your agent returns incoherent responses because the model was swapped

Traditional health checks don't catch any of this. **agent-probe** does.

## What It Checks

| Check | What It Does |
|-------|-------------|
| **LLM Connectivity** | Pings your LLM endpoint, verifies API key validity |
| **Token Budget** | Queries remaining tokens/credits against configurable thresholds |
| **Tool Availability** | Verifies external tool endpoints your agent depends on |
| **Response Quality** | Sends a test prompt, validates the response is coherent |

## Installation

### From Source

```bash
go install github.com/arcane-bear/agent-probe/cmd/agent-probe@latest
agent-probe --config agent-probe.yaml
```

### Docker

```bash
docker run -d \
  -p 8089:8089 \
  -v $(pwd)/agent-probe.yaml:/etc/agent-probe/config.yaml \
  ghcr.io/arcane-bear/agent-probe:latest
```

### Kubernetes Sidecar

Add agent-probe as a sidecar container to your agent pod:

```yaml
containers:
  - name: my-agent
    image: my-agent:latest
    ports:
      - containerPort: 8080

  - name: agent-probe
    image: ghcr.io/arcane-bear/agent-probe:latest
    ports:
      - containerPort: 8089
    volumeMounts:
      - name: probe-config
        mountPath: /etc/agent-probe

volumes:
  - name: probe-config
    configMap:
      name: agent-probe-config
```

Then point your liveness/readiness probes at agent-probe:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8089
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /readyz
    port: 8089
  initialDelaySeconds: 5
  periodSeconds: 15
```

## Usage

```bash
# Run with a config file
agent-probe --config agent-probe.yaml

# Check the health endpoint
curl http://localhost:8089/healthz

# Check readiness
curl http://localhost:8089/readyz
```

## Configuration

See [agent-probe.yaml](agent-probe.yaml) for a full example. Here's a minimal config:

```yaml
server:
  port: 8089

checks:
  llm:
    enabled: true
    endpoint: "https://api.openai.com/v1/chat/completions"
    api_key_env: "OPENAI_API_KEY"
    timeout: 10s

  token_budget:
    enabled: true
    endpoint: "https://api.openai.com/v1/usage"
    api_key_env: "OPENAI_API_KEY"
    warn_threshold: 20
    critical_threshold: 5

  tools:
    enabled: true
    endpoints:
      - name: "vector-db"
        url: "http://qdrant:6333/healthz"
        timeout: 5s

  response_quality:
    enabled: false
    endpoint: "https://api.openai.com/v1/chat/completions"
    api_key_env: "OPENAI_API_KEY"
    test_prompt: "What is 2+2?"
    expected_substring: "4"
    timeout: 15s
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /healthz` | Full health check — runs all enabled checks |
| `GET /readyz` | Readiness check — lighter weight, skips expensive checks |
| `GET /healthz/{check}` | Run a specific check (e.g., `/healthz/llm`) |
| `GET /metrics` | Prometheus-compatible metrics |

## Response Format

```json
{
  "status": "degraded",
  "checks": {
    "llm": { "status": "healthy", "latency_ms": 142 },
    "token_budget": { "status": "warning", "remaining_pct": 12.5 },
    "tools": {
      "status": "unhealthy",
      "details": {
        "vector-db": { "status": "unhealthy", "error": "connection refused" }
      }
    },
    "response_quality": { "status": "healthy", "latency_ms": 980 }
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## Documentation

For deployment guides and integration with other RapidClaw tools, visit the [RapidClaw docs](https://rapidclaw.dev).

## Related Projects

- [RapidClaw](https://rapidclaw.dev) — self-hosted AI agent deployment platform
- [arcane-bear](https://github.com/arcane-bear) — more open-source tools for AI infrastructure

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](LICENSE).

---

Built by [RapidClaw](https://rapidclaw.dev) | [arcane-bear](https://github.com/arcane-bear)
