# Historical Data API

A production-ready RESTful API service built with Go for managing historical financial data.

## 🚀 Features

- **High Performance**: Fiber v2 framework
- **Clean Architecture**: Clear separation of concerns (Controller → Service → Repository)
- **Database**: MySQL 8.0+ with GORM
- **CSV Upload**: Streaming CSV parser with batch processing (1000 records/batch)
- **Structured Logging**: Zerolog for efficient logging
- **Validation**: Request validation with go-playground/validator
- **Rate Limiting**: IP-based rate limiting
- **Containerization**: Docker & Docker Compose
- **CI/CD**: Jenkins
- **Production Ready**: Health checks, graceful shutdown, error handling

## 📋 Prerequisites
- Go 1.21 or higher
- Docker & Docker Compose

## 🏗️ Project Structure

```
go-historical-data/
├── cmd/api/                    # Application entry point
├── internal/                   # Private application code
│   ├── controller/             # HTTP handlers
│   ├── service/                # Business logic
│   ├── repository/             # Data access layer
│   ├── model/                  # Domain models
│   ├── dto/                    # Data Transfer Objects
│   └── middleware/             # HTTP middleware
├── pkg/                        # Public reusable packages
│   ├── config/                 # Configuration management
│   ├── database/               # Database connections
│   ├── logger/                 # Logging utilities
│   ├── validator/              # Validation utilities
│   ├── csvparser/              # CSV parsing utilities
│   └── response/               # Response helpers
├── database/migrations/        # SQL migrations
├── config/                     # Configuration files
└── docker-compose.yml          # Docker services configuration
```

## 🚦 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd go-historical-data
```

### 2. Start MySQL and go-historical-data app

```bash
docker-compose up
```

The API will be available at `http://localhost:8080`

## 📚 API Endpoints

### Health Check
- `GET /health` - Application health status

### Historical Data
- `POST /api/v1/data` - Upload historical data (JSON bulk)
- `GET /api/v1/data` - Retrieve historical data with filters
- `GET /api/v1/data/:id` - Get specific historical data by ID
