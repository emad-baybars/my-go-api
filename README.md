# Go Backend API Template

A comprehensive, production-ready Go backend template with authentication, authorization, localization, and database support for both MongoDB and PostgreSQL.

## 🚀 Features

- **RESTful API** with Gin framework
- **Authentication & Authorization** with JWT tokens and role-based access control
- **Multi-language Support** (English, Arabic, German) with localization
- **Database Support** for both MongoDB and PostgreSQL with GORM
- **API Documentation** with Swagger/OpenAPI
- **Comprehensive Middleware** (CORS, Rate Limiting, Logging, Recovery)
- **Docker Support** with multi-stage builds
- **Error Handling** with structured logging
- **Health Checks** for monitoring
- **Input Validation** and sanitization
- **Security Best Practices** implemented

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/download/) (version 1.21 or higher)
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)
- [Make](https://www.gnu.org/software/make/) (optional, for using Makefile commands)

### Optional Development Tools

- [golangci-lint](https://golangci-lint.run/usage/install/) for code linting
- [swag](https://github.com/swaggo/swag) for generating Swagger documentation
- [migrate](https://github.com/golang-migrate/migrate) for database migrations
- [Air](https://github.com/cosmtrek/air) for live reloading during development

## 🛠️ Project Setup

### Step 1: Clone or Create the Project

```bash
# Create a new directory for your project
mkdir my-go-api
cd my-go-api

# Initialize git repository
git init

# Create the Go module
go mod init my-go-api
```

### Step 2: Create Project Structure

Create the following directory structure:

```
my-go-api/
├── config/
│   └── config.go
├── database/
│   └── database.go
├── handlers/
│   └── handlers.go
├── middleware/
│   └── middleware.go
├── models/
│   └── models.go
├── routes/
│   └── routes.go
├── utils/
│   └── utils.go
├── docs/
├── scripts/
├── nginx/
├── logs/
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
├── Makefile
└── README.md
```

### Step 3: Copy the Source Code

Copy all the provided source code files into their respective directories:

1. Copy `main.go` to the root directory
2. Copy `config/config.go`
3. Copy `database/database.go`
4. Copy `models/models.go`
5. Copy `middleware/middleware.go`
6. Copy `utils/utils.go`
7. Copy `handlers/handlers.go`
8. Copy `routes/routes.go`
9. Copy `go.mod`
10. Copy `Dockerfile`
11. Copy `docker-compose.yml`
12. Copy `.env.example`
13. Copy `Makefile`

### Step 4: Install Dependencies

```bash
# Download and install all dependencies
go mod download
go mod tidy

# Install development tools (optional)
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/cosmtrek/air@latest
```

### Step 5: Create Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit the .env file with your configuration
nano .env  # or use your preferred editor
```

Update the `.env` file with your database credentials and other configuration:

```env
# Application Configuration
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=info
DEFAULT_LANGUAGE=en

# JWT Configuration (IMPORTANT: Change this in production!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# PostgreSQL Database Configuration
POSTGRES_ENABLED=true
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USERNAME=postgres
POSTGRES_PASSWORD=your-password
POSTGRES_DATABASE=your-database
POSTGRES_SSLMODE=disable

# MongoDB Database Configuration (optional)
MONGODB_ENABLED=false
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_DATABASE=your-database
```

### Step 6: Create Additional Required Files

#### Create `.gitignore`:

```bash
cat > .gitignore << 'EOF'
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
build/

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.html

# Go workspace file
go.work

# Environment variables
.env
.env.local
.env.development
.env.production

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Logs
logs/
*.log

# Dependencies
vendor/

# Swagger generated files
docs/

# Docker volumes
data/

# Temporary files
tmp/
temp/
EOF
```

#### Create database initialization scripts:

```bash
# Create scripts directory
mkdir -p scripts

# PostgreSQL initialization script
cat > scripts/init-postgres.sql << 'EOF'
-- Create database if not exists
-- This script runs when PostgreSQL container starts

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- You can add more initialization SQL here
EOF

# MongoDB initialization script
cat > scripts/init-mongo.js << 'EOF'
// MongoDB initialization script
// This script runs when MongoDB container starts

db = db.getSiblingDB('backend_template');

// Create collections and indexes
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "username": 1 }, { unique: true });

// You can add more initialization here
EOF
```

#### Create Nginx configuration:

```bash
# Create nginx directory and configuration
mkdir -p nginx

cat > nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream api {
        server api:8080;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
EOF
```

### Step 7: Generate Swagger Documentation

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
swag init -g main.go -o docs/
```

### Step 8: Set Up Development Tools (Optional)

#### Create Air configuration for live reloading:

```bash
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
args_bin = []
bin = "./tmp/main"
cmd = "go build -o ./tmp/main ."
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "testdata"]
exclude_file = []
exclude_regex = ["_test.go"]
exclude_unchanged = false
follow_symlink = false
full_bin = ""
include_dir = []
include_ext = ["go", "tpl", "tmpl", "html"]
kill_delay = "0s"
log = "build-errors.log"
send_interrupt = false
stop_on_root = false

[color]
app = ""
build = "yellow"
main = "magenta"
runner = "green"
watcher = "cyan"

[log]
time = false

[misc]
clean_on_exit = false

[screen]
clear_on_rebuild = false
EOF
```

## 🚀 Running the Application

### Option 1: Local Development

1. **Start databases locally** (if not using Docker):
   ```bash
   # Start PostgreSQL (example with Homebrew on macOS)
   brew services start postgresql
   
   # Start MongoDB (example with Homebrew on macOS)
   brew services start mongodb-community
   ```

2. **Run the application**:
   ```bash
   # Using Go directly
   go run main.go
   
   # Or using Air for live reloading
   air
   
   # Or using Make
   make run
   ```

### Option 2: Docker Development

1. **Start all services with Docker Compose**:
   ```bash
   # Start all services (API, PostgreSQL, MongoDB, Redis, Nginx)
   docker-compose up -d
   
   # View logs
   docker-compose logs -f api
   
   # Or using Make
   make compose-up
   ```

2. **Check service status**:
   ```bash
   docker-compose ps
   ```

### Option 3: Docker API Only

```bash
# Build the Docker image
docker build -t my-go-api .

# Run the container
docker run -p 8080:8080 --env-file .env my-go-api

# Or using Make
make docker-build
make docker-run
```

## 📚 API Usage

Once the application is running, you can access:

- **API Base URL**: `http://localhost:8080/api/v1`
- **Swagger Documentation**: `http://localhost:8080/swagger/index.html`
- **Health Check**: `http://localhost:8080/api/v1/health`

### Example API Calls

#### 1. Health Check
```bash
curl -X GET http://localhost:8080/api/v1/health
```

#### 2. User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### 3. User Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### 4. Get User Profile (requires authentication)
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 5. Update User Profile
```bash
curl -X PUT http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Updated John",
    "last_name": "Updated Doe"
  }'
```

## 🔧 Development Workflow

### Using Make Commands

```bash
# Setup development environment
make setup

# Run with live reloading
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make format

# Generate Swagger docs
make swagger

# Build for production
make build

# Run CI pipeline
make ci

# Docker commands
make docker-build
make docker-run
make docker-stop

# Docker Compose commands
make compose-up
make compose-down
make compose-logs
```

### Database Migrations (PostgreSQL)

```bash
# Create a new migration
make db-create-migration NAME=create_users_table

# Run migrations
make db-migrate-up

# Rollback migrations
make db-migrate-down
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Check for security vulnerabilities
gosec ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

## 🌍 Localization

The API supports multiple languages. Set the `Accept-Language` header or `lang` query parameter:

```bash
# English (default)
curl -X GET http://localhost:8080/api/v1/health

# Arabic
curl -X GET http://localhost:8080/api/v1/health \
  -H "Accept-Language: ar"

# German
curl -X GET http://localhost:8080/api/v1/health \
  -H "Accept-Language: de"

# Using query parameter
curl -X GET "http://localhost:8080/api/v1/health?lang=ar"
```

## 🔒 Security Features

- **JWT Authentication** with configurable expiration
- **Role-based Authorization** (user, admin, superadmin)
- **Password Hashing** with bcrypt
- **Rate Limiting** to prevent abuse
- **CORS Protection** with configurable origins
- **Request ID Tracking** for debugging
- **Input Validation** and sanitization
- **Secure Headers** and HTTPS support

## 🐳 Docker Configuration

### Environment Variables in Docker

The Docker setup includes:
- **Multi-stage builds** for optimized image size
- **Non-root user** for security
- **Health checks** for monitoring
- **Volume mounts** for persistence
- **Network isolation** between services

### Production Deployment

```bash
# Build for production
docker build -t my-go-api:prod .

# Run with production environment
docker run -d \
  --name my-go-api-prod \
  -p 8080:8080 \
  -e ENVIRONMENT=production \
  -e JWT_SECRET=your-production-secret \
  my-go-api:prod
```

## 📊 Monitoring and Logging

### Health Checks

```bash
# Check application health
curl http://localhost:8080/api/v1/health

# Expected response
{
  "success": true,
  "message": "System is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "services": {
      "postgresql": "healthy",
      "mongodb": "healthy"
    },
    "version": "1.0.0"
  }
}
```

### Logs

Logs are structured in JSON format and include:
- Request ID for tracing
- Timestamp and log level
- HTTP request details
- Error information
- Performance metrics

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./handlers -v

# Run benchmarks
go test -bench=. ./...
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENVIRONMENT` | Application environment | `development` | No |
| `PORT` | Server port | `8080` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `POSTGRES_ENABLED` | Enable PostgreSQL | `true` | No |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` | No |
| `POSTGRES_PORT` | PostgreSQL port | `5432` | No |
| `POSTGRES_USERNAME` | PostgreSQL username | `postgres` | No |
| `POSTGRES_PASSWORD` | PostgreSQL password | - | Yes if enabled |
| `MONGODB_ENABLED` | Enable MongoDB | `false` | No |
| `MONGODB_HOST` | MongoDB host | `localhost` | No |
| `MONGODB_PORT` | MongoDB port | `27017` | No |

## 🚀 Deployment

### Production Checklist

- [ ] Update `JWT_SECRET` with a strong secret
- [ ] Set `ENVIRONMENT=production`
- [ ] Configure proper database credentials
- [ ] Set up SSL certificates
- [ ] Configure monitoring and alerting
- [ ] Set up backup strategies
- [ ] Review and configure CORS settings
- [ ] Set up log aggregation
- [ ] Configure rate limiting for production traffic

### Deployment Options

1. **Docker Container**
2. **Kubernetes**
3. **Cloud Platforms** (AWS, GCP, Azure)
4. **Traditional Servers**

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the full test suite
6. Submit a pull request

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Common Issues

1. **Port already in use**:
   ```bash
   # Kill process using port 8080
   lsof -ti:8080 | xargs kill -9
   ```

2. **Database connection failed**:
   - Check database credentials in `.env`
   - Ensure database service is running
   - Verify network connectivity

3. **Docker build fails**:
   - Check Docker daemon is running
   - Verify Dockerfile syntax
   - Clear Docker cache: `docker system prune`

4. **Swagger not generating**:
   ```bash
   # Install swag
   go install github.com/swaggo/swag/cmd/swag@latest
   
   # Regenerate docs
   swag init -g main.go -o docs/
   ```

### Getting Help

- Check the [Issues](../../issues) section
- Review the [Swagger Documentation](http://localhost:8080/swagger/index.html)
- Enable debug logging: `LOG_LEVEL=debug`

## 📚 Additional Resources

- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [MongoDB Go Driver](https://docs.mongodb.com/drivers/go/)
- [JWT Go Library](https://github.com/golang-jwt/jwt)
- [Docker Documentation](https://docs.docker.com/)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
