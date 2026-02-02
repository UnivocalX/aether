# Aether

A data management platform for organizing and tagging data stored in Object Storage with advanced support for datasets and AI-powered automatic tag generation.

## Overview

Aether integrates with S3-compatible storage and PostgreSQL to provide:

- **Intelligent Tagging** - Efficiently organize data with automatic and manual tags
- **Fast Retrieval** - Enable quick querying through optimized metadata indexing
- **Deduplication** - Prevent storage waste by identifying duplicate assets via SHA256 hashing
- **Dataset Management** - Group related assets for better organization
- **AI Integration** - Automatic tag generation powered by AI agents

Aether simplifies data governance by combining automation, scalability, and intelligent metadata handling, making it ideal for teams managing large-scale unstructured data.

## Features

- RESTful API for asset and tag management
- S3-compatible object storage integration (MinIO, AWS S3, etc.)
- PostgreSQL-backed metadata registry
- SHA256-based deduplication
- Presigned URL generation for secure asset access
- Comprehensive tag system for flexible categorization
- CLI for asset management operations

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- MinIO or S3-compatible storage
- PostgreSQL 14+

## Installation

### From Source

```bash
go build -o bin/aether ./cmd/
```

### Using GoReleaser

Build for all platforms:
```bash
goreleaser build --clean
```

Build for specific platform:
```bash
export GOOS="linux"
export GOARCH="amd64"
goreleaser build --single-target --clean
```

## Configuration

### 1. Environment Setup

Create a `.env` file in the project root:

```env
# MinIO/S3 Configuration
MINIO_ROOT_USER=admin
MINIO_ROOT_PASSWORD=your_secure_password

# Object Storage
MINIO_BUCKET=aether-production

# PostgreSQL Configuration
POSTGRES_USER=admin
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=aether

# PGAdmin (Optional)
PGADMIN_DEFAULT_EMAIL=admin@example.com
PGADMIN_DEFAULT_PASSWORD=your_secure_password
PGADMIN_CONFIG_SERVER_MODE=False

# Client Credentials (Configure after creating MinIO access keys)
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
```

### 2. Application Configuration

Create or edit `$HOME/.aether/config.yaml`:

```yaml
# Logging Level
level: debug

# API Endpoint (for CLI client)
endpoint: localhost:9090

# Server Configuration
server:
  port: 9090
  production: false

  # S3-Compatible Storage
  storage:
    s3endpoint: "http://localhost:9000"
    bucket: "aether-production"
    prefix: "aether/assets"

  # Database Connection
  database:
    endpoint: "localhost:5432"
    user: "admin"
    password: "your_secure_password"
    name: "aether"
    ssl: false
```

## Quick Start

### 1. Start Infrastructure Services

```bash
docker compose up -d
```

This will start:
- MinIO (Object Storage) on port 9000
- PostgreSQL (Database) on port 5432
- PGAdmin (Optional) on port 5050

### 2. Create MinIO Access Keys

1. Open MinIO Console at http://localhost:9000
2. Log in with credentials from `.env` file
3. Navigate to **Access Keys** → **Create Access Key**
4. Copy the generated credentials
5. Update `.env` file with `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

### 3. Start Aether Server

```bash
./bin/aether serve
```

The API server will start on http://localhost:9090

### 4. Verify Installation

```bash
curl http://localhost:9090/health
```

## Usage

### CLI Commands

#### Start Server
```bash
aether serve
```

#### Load Assets
```bash
aether assets load /path/to/files
```

## API Documentation

Import the Postman collection for interactive API documentation:

1. Open Postman
2. Import `docs/collection.json`
3. Import `docs/environment.json`
4. Select "Aether - Local" environment
5. Start making requests

See [ROADMAP.md](ROADMAP.md) for planned features and improvements.

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development guidelines.

## License

See [LICENSE](LICENSE) for details.