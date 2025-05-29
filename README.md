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
OR:
```bash
docker-compose up -d

TELEGRAM_TOKEN=t DB_USER=postgres DB_PASSWORD=postgres DB_NAME=adtime DB_HOST=localhost REDIS_ADDR=localhost:6379 make run

TELEGRAM_TOKEN=t DB_USER=postgres DB_PASSWORD=postgres DB_NAME=adtime make run

TELEGRAM_TOKEN=t make run
```

# Start services
docker-compose up -d redis postgres

# Run migrations (choose one)
make migrate-up
# OR
goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime host=localhost sslmode=disable" up

# Run the bot
```bash
TELEGRAM_TOKEN=t DB_USER=postgres DB_PASSWORD=postgres DB_NAME=adtime DB_HOST=localhost REDIS_ADDR=localhost:6379 make run

goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime host=localhost sslmode=disable" status

docker-compose exec postgres psql -U postgres -d adtime -c "\dt"

TELEGRAM_TOKEN=t DB_USER=postgres DB_PASSWORD=postgres DB_NAME=adtime DB_HOST=localhost REDIS_ADDR=localhost:6379 make run

TELEGRAM_TOKEN=t \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=adtime \
DB_HOST=localhost \
REDIS_ADDR=localhost:6379 \
make run

docker-compose exec postgres psql -U postgres -d adtime -c "SELECT * FROM textures;"

docker-compose exec postgres psql -U postgres -d adtime -c "
INSERT INTO textures (id, name, price_per_dm2, in_stock) VALUES 
('11111111-1111-1111-1111-111111111111', 'Стандартная текстура', 10.0, true),
('22222222-2222-2222-2222-222222222222', 'Премиум текстура', 15.5, true);"
```

### HOW TO CHECK INFO

Connect to your PostgreSQL container:
```bash
docker-compose exec postgres psql -U postgres -d adtime
```

List all tables to confirm structure (you've already done this):
```bash
\dt
```

Query the orders table to see all orders:
```sql
SELECT * FROM orders;
```

For more specific information about the user's order (user ID 5756911009):
```sql
SELECT * FROM orders WHERE user_id = 5756911009;
```

If you want to see the order with ID 1:
```sql
SELECT * FROM orders WHERE id = 1;
```