# AgentRuntime - Quick Start Guide

Get AgentRuntime up and running in under 5 minutes!

## Prerequisites

- **Go 1.22+** - [Install Go](https://go.dev/doc/install)
- **Node.js 18+** - [Install Node](https://nodejs.org/) (for Web UI)
- **Git** - [Install Git](https://git-scm.com/)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/NikoSokratous/agentctl.git
cd agentruntime
```

### 2. Build the Project

```bash
# Build all binaries
make build

# Or build individually
go build -o bin/agentctl ./cmd/agentctl
go build -o bin/server ./pkg/orchestrate/cmd/server
```

### 3. Initialize Database

```bash
# Run migrations
./bin/agentctl init

# Or manually with sqlite3
cat migrations/*.sql | sqlite3 agentruntime.db
```

## Running AgentRuntime

### Option 1: Local Development (Recommended for testing)

```bash
# Start the server
./bin/server

# In another terminal, use the CLI
./bin/agentctl --help
```

### Option 2: Docker

```bash
# Build and run with Docker
docker build -t agentruntime .
docker run -p 8080:8080 -p 3000:3000 agentruntime
```

### Option 3: Kubernetes

```bash
# Deploy with Helm
helm install agentruntime ./k8s/helm \
  --set postgresql.enabled=true \
  --set redis.enabled=true
```

## Your First Agent

### 1. Configure LLM Provider

```bash
# Set API key (choose one)
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
# Or use Ollama (no key needed)
export LLM_PROVIDER="ollama"
```

### 2. Create an Agent

```bash
./bin/agentctl agent create my-first-agent \
  --goal "Analyze the latest AI research papers" \
  --llm gpt-4 \
  --max-steps 10 \
  --autonomy standard
```

### 3. Run the Agent

```bash
# Start execution
./bin/agentctl agent run my-first-agent

# Watch live
./bin/agentctl runs watch <run-id>

# Get results
./bin/agentctl runs get <run-id>
```

## Using the Web UI

### 1. Install Dependencies

```bash
cd web
npm install
```

### 2. Start Development Server

```bash
npm run dev
```

### 3. Open Browser

Navigate to [http://localhost:3000](http://localhost:3000)

**Available Pages:**
- **Dashboard** - Overview and recent runs
- **Workflow Designer** - Visual workflow builder
- **Marketplace** - Pre-built workflow templates
- **Analytics** - Cost and performance metrics

## Create Your First Workflow

### Using the CLI

```yaml
# workflow.yaml
name: "research-workflow"
description: "Research and summarize AI papers"
steps:
  - name: "search"
    agent: "researcher"
    goal: "Find papers on {{topic}}"
    output_key: "papers"
    
  - name: "analyze"
    agent: "analyzer"
    goal: "Analyze papers: {{.Outputs.papers}}"
    output_key: "analysis"
    
  - name: "summarize"
    agent: "writer"
    goal: "Create summary from {{.Outputs.analysis}}"
```

```bash
# Run the workflow
./bin/agentctl workflow run workflow.yaml \
  --param topic="AI safety"
```

### Using the Web UI

1. Go to **Workflow Designer**
2. Drag agents onto the canvas
3. Connect them to create a flow
4. Configure each step
5. Click **Run** to execute

## Add a Knowledge Base (RAG)

Ground your agent in your documentation using RAG (Retrieval Augmented Generation).

### 1. Configure Embeddings

```bash
# Set OpenAI API key (or use local embeddings - see docs/EMBEDDINGS.md)
export OPENAI_API_KEY="sk-..."
```

### 2. Ingest Documents

From the project root:

```bash
# Ingest markdown and text files from the knowledge-base example
./bin/agentctl context ingest examples/knowledge-base/docs --source "documentation"

# Verify
./bin/agentctl context knowledge list
```

### 3. Run the Agent

```bash
./bin/agentctl run --agent examples/knowledge-base/agent.yaml --goal "How do I deploy an agent?"
```

The agent retrieves relevant chunks from your docs and grounds its response in them.

### 4. Test Search

```bash
./bin/agentctl context search "how do I configure policies?" --top-k 5
```

See [examples/knowledge-base/README.md](examples/knowledge-base/README.md) and [docs/EMBEDDINGS.md](docs/EMBEDDINGS.md) for details.

---

## Working with Policies

### Create a Policy

```yaml
# policies/safety.yaml
name: security-policy
version: "1.0.0"
description: "Prevent dangerous operations"
environment: production

rules:
  - id: block-file-deletion
    description: "Block file deletion"
    condition: |
      tool.name == "file_delete"
    action: deny
    severity: high
```

### Apply Policy

```bash
./bin/agentctl policy apply policies/safety.yaml
```

### Test Policy

```bash
./bin/agentctl policy simulate security-policy 1.0.0 \
  --run-id <past-run-id>
```

## Exploring Templates

### List Available Templates

```bash
./bin/agentctl workflow templates list
```

### Install a Template

```bash
# Install from marketplace
./bin/agentctl workflow templates install code-review

# Customize and run
./bin/agentctl workflow run code-review \
  --param repository_url=https://github.com/user/repo \
  --param branch=main
```

### Popular Templates

- **knowledge-base** - Support agent with RAG over your docs
- **code-review** - Automated code review with linting and security
- **data-pipeline** - ETL workflow with validation
- **research** - Multi-source research and synthesis
- **customer-support** - Ticket routing and responses
- **content-creation** - SEO-optimized content generation
- **devops-deployment** - Deployment with testing and rollback

## Development Tools

### Validate Tools

```bash
# Validate a tool definition
./bin/agentctl tool validate tools/my-tool.yaml
```

### Scaffold New Tool

```bash
# Generate tool boilerplate
./bin/agentctl scaffold tool --name my_tool --output tools/
```

### Debug Workflows

```bash
# Start debug session
./bin/agentctl workflow debug <workflow-id>

# Set breakpoint
> break step-name

# Step through
> step

# Inspect variables
> inspect
```

## Monitoring & Observability

### View Metrics

```bash
# Prometheus metrics
curl http://localhost:8080/metrics

# Health check
curl http://localhost:8080/health
```

### Enable Tracing

```bash
# Start Jaeger
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one

# Configure AgentRuntime
export OTLP_ENDPOINT="localhost:4317"
export TRACING_ENABLED="true"

# View traces
open http://localhost:16686
```

### Cost Tracking

```bash
# View costs
./bin/agentctl costs --by agent
./bin/agentctl costs --by tenant
./bin/agentctl costs --date-range "2024-01-01:2024-01-31"
```

## Next Steps

### Learn More
- 📖 Read the [User Guide](docs/USER_GUIDE.md)
- 🏗️ Understand the [Architecture](docs/ARCHITECTURE.md)
- 🧠 Add [RAG and Embeddings](docs/EMBEDDINGS.md)
- 🔌 Build [Custom Tools](docs/PLUGIN_DEVELOPMENT.md)
- 🚀 Deploy to [Production](docs/DEPLOYMENT.md)

### Examples
- Check out [examples/](examples/) for 20+ ready-to-use examples
- Browse [Workflow Templates](examples/workflows/templates/README.md)

### Community
- 💬 Join [Discord](https://discord.gg/agentruntime)
- 🐛 Report issues on [GitHub](https://github.com/NikoSokratous/agentctl/issues)
- ⭐ Star the [repository](https://github.com/NikoSokratous/agentctl)

## Common Issues

### Database Not Found
```bash
# Initialize database
./bin/agentctl init
```

### Permission Denied
```bash
# Make binaries executable
chmod +x bin/*
```

### Port Already in Use
```bash
# Change port
export PORT=8081
./bin/server
```

### LLM API Key Not Set
```bash
# Set your API key
export OPENAI_API_KEY="sk-..."
```

## Configuration

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for all configuration options.

Key environment variables:
- `OPENAI_API_KEY` - OpenAI API key (also used for embeddings when using RAG)
- `ANTHROPIC_API_KEY` - Anthropic API key
- `DATABASE_URL` - Database connection string
- `REDIS_URL` - Redis connection string
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

---

**Need help?** Open an issue or join our Discord!
