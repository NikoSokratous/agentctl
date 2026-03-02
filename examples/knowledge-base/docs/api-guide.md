# API Guide

## Authentication

To use the AgentRuntime API, you need to authenticate with an API key.

### Getting an API Key

1. Sign up for an account at https://agentruntime.io
2. Navigate to Settings > API Keys
3. Click "Generate New API Key"
4. Copy the key and store it securely

### Using Your API Key

Set the API key as an environment variable:

```bash
export AGENTRUNTIME_API_KEY="ar_..."
```

Or pass it directly in your code:

```go
client := agentruntime.NewClient(agentruntime.Config{
    APIKey: "ar_...",
})
```

## Making API Calls

### Create an Agent

```bash
curl -X POST https://api.agentruntime.io/v1/agents \
  -H "Authorization: Bearer $AGENTRUNTIME_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my_agent",
    "model": "gpt-4",
    "tools": ["search", "calculator"]
  }'
```

### Run an Agent

```bash
curl -X POST https://api.agentruntime.io/v1/agents/my_agent/run \
  -H "Authorization: Bearer $AGENTRUNTIME_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "goal": "Calculate the sum of 42 and 17"
  }'
```

### Check Run Status

```bash
curl https://api.agentruntime.io/v1/runs/run_123 \
  -H "Authorization: Bearer $AGENTRUNTIME_API_KEY"
```

## Rate Limits

- Free tier: 100 requests per hour
- Pro tier: 1000 requests per hour
- Enterprise: Custom limits

## Error Handling

The API uses standard HTTP status codes:

- 200: Success
- 400: Bad Request (invalid parameters)
- 401: Unauthorized (missing or invalid API key)
- 403: Forbidden (insufficient permissions)
- 429: Too Many Requests (rate limit exceeded)
- 500: Internal Server Error

## Best Practices

1. **Store API keys securely**: Never commit keys to version control
2. **Use environment variables**: Keep keys out of code
3. **Handle rate limits**: Implement exponential backoff
4. **Validate responses**: Check status codes and error messages
5. **Use webhooks**: For long-running operations, use webhooks instead of polling

## SDKs

We provide official SDKs for:

- Go: `go get github.com/agentruntime/agentruntime-go`
- Python: `pip install agentruntime`
- JavaScript: `npm install @agentruntime/client`

Example in Go:

```go
import "github.com/agentruntime/agentruntime-go"

client := agentruntime.NewClient(agentruntime.Config{
    APIKey: os.Getenv("AGENTRUNTIME_API_KEY"),
})

agent, err := client.CreateAgent(ctx, &agentruntime.AgentConfig{
    Name: "support_agent",
    Model: "gpt-4",
})
```

## Support

For API questions:
- Email: api-support@agentruntime.io
- Discord: https://discord.gg/agentruntime
- Documentation: https://docs.agentruntime.io
