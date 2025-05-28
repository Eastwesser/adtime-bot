# ADTIME Telegram Bot (Go)

A Telegram bot for managing orders with PostgreSQL storage and Redis for state management.

## Project Structure
```bash
    adtime-bot/
    ├── cmd/                  # Main application entry points
    ├── internal/             # Internal application code
    │   ├── bot/              # Telegram bot logic
    │   ├── config/           # Configuration handling
    │   └── storage/          # Database storage implementation
    ├── pkg/                  # Reusable packages
    │   ├── api/              # API client
    │   ├── logger/           # Logging utilities
    │   └── redis/            # Redis client
    ├── migrations/           # Database migrations
    ├── Dockerfile            # Production Dockerfile
    ├── docker-compose.yml    # Development environment
    └── Makefile              # Development tasks
```

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- PostgreSQL
- Redis
- Telegram Bot Token

## Getting Started

### 1. Environment Setup

Create a `.env` file in the project root with the following variables:

```bash
    TELEGRAM_TOKEN=your_telegram_bot_token
    API_BASE_URL=your_api_base_url
    API_KEY=your_api_key
    DB_USER=postgres
    DB_PASSWORD=postgres
    DB_NAME=adtime
```

### 2. Installation
```bash
# Clone the repository
git clone https://github.com/yourusername/adtime-bot.git
cd adtime-bot

# Install dependencies
go mod download
```

### 3. Running the Application
Option 1: Local Development
```bash

# Build and run
make build
./bin/adtime-bot

# Or run directly
make run
```
Option 2: Docker
```bash

# Build and start containers
make docker-build
make docker-up

# To stop
make docker-down
```

### 4. Database Migrations
```bash
# Apply all migrations
make migrate-up

# Create a new migration
make migrate-create
# Follow the prompt to enter migration name

# Check migration status
make migrate-status

# Revert the last migration
make migrate-down
```

## Available Make Commands

Run make help to see all available commands:
```bash
make help
```

## Testing

To run tests:
```bash
make test
```

## Deployment

Build the Docker image:
```bash
make docker-build
```

Start the services:
```bash
make docker-up
```
